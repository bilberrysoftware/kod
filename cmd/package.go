/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"fmt"
	"kod/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var chartPath string

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Packages an individual Helm chart or helmfile",
	Long: `Creates a deployable package archive consistent of the helm chart or helm file
and all of its constituent container images.`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO sanity check parameter values

		// See if the Chart.yaml file exists, error if not
		folder, err := os.Stat(chartPath)
		if err != nil {
			fmt.Println(chartPath, "does not exist")
			os.Exit(1)
		}
		if !folder.IsDir() {
			fmt.Println("Chart folder", chartPath, "is not a directory")
			os.Exit(1)
		}
		chartYaml := filepath.Join(chartPath, "Chart.yaml")
		chartYamlFile, err := os.Stat(chartYaml)
		if err != nil {
			fmt.Println(chartYaml, "does not exist")
			os.Exit(1)
		}
		if chartYamlFile.IsDir() {
			fmt.Println(chartYaml, "is a directory and not a YAML file")
			os.Exit(1)
		}
		// Now read the YAML, and save the name and version information
		fileBytes, err := os.ReadFile(chartYaml)
		if err != nil {
			fmt.Println("File cannot be read", err)
			os.Exit(1)
		}
		var chartDef = types.HelmChart{}
		err = yaml.Unmarshal(fileBytes, &chartDef)
		if err != nil {
			fmt.Println("Error parsing Helm Chart definition", err)
			os.Exit(1)
		}
		// Now print what we've found
		fmt.Println(fmt.Sprintf("Packaging chart:: name: '%s', version: '%s', appVersion: `%s`",
			chartDef.Name, chartDef.Version, chartDef.AppVersion))
		// Create temporary folder based on chart name and version
		chartAndVersion := fmt.Sprintf("%s-%s", chartDef.Name, chartDef.Version)
		folderName := "kod-" + chartAndVersion
		tempPath := filepath.Join(os.TempDir(), folderName)
		err = os.MkdirAll(tempPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating temp folder", err)
			os.Exit(1)
		}

		// Determine the container images required
		var containers []types.ContainerImage
		// Read the values YAML file and look for common properties
		valuesPath := filepath.Join(chartPath, "values.yaml")
		valuesFile, err := os.Stat(valuesPath)
		if err != nil {
			fmt.Println(valuesPath, "does not exist")
			os.Exit(1)
		}
		if valuesFile.IsDir() {
			fmt.Println(valuesPath, "is a directory and not a YAML file")
			os.Exit(1)
		}
		valuesMap := make(map[string]interface{})
		valuesYaml, err := os.ReadFile(valuesPath)
		if err != nil {
			fmt.Println(valuesYaml, "could not be read", err)
			os.Exit(1)
		}
		err = yaml.Unmarshal(valuesYaml, &valuesMap)
		if err != nil {
			fmt.Println(valuesPath, "could not be parsed", err)
			os.Exit(1)
		}
		// Now look for various common image config locations
		vImageEl := valuesMap["image"]
		if vImageEl != nil {
			vImage := vImageEl.(map[string]interface{})
			// check for 'registry' first
			vImageRegistry := vImage["registry"]
			vImageRepository := vImage["repository"].(string)
			vImageRegistryStr := ""
			// TODO trim strings of whitespace
			if vImageRepository != "" {
				vImageDigest := vImage["digest"]
				vImageDigestStr := ""
				if vImageDigest == nil {
					// take the last part from repository after the @, if specified, otherwise leave as ""
					idx := strings.LastIndex(vImageRepository, "@")
					if idx != -1 {
						vImageRepository = vImageRepository[:idx]
						vImageDigestStr = vImageRepository[idx+1:]
					}
				} else {
					vImageDigestStr = vImageDigest.(string)
				}
				if vImageRegistry == nil {
					// Get Registry from the first part of repository if it looks like a URL, OR default to docker.io
					idx := strings.Index(vImageRepository, "/")
					if idx != -1 {
						vImageRegistryStr = vImageRepository[:idx]
						vImageRepository = vImageRepository[idx+1:]
					}
				} else {
					vImageRegistryStr = vImageRegistry.(string)
				}
				vImageTag := vImage["tag"]
				vImageTagStr := ""
				if vImageTag == nil {
					// Get Tag from the last part of the repository (now that we've removed digest
					idx := strings.LastIndex(vImageRepository, ":")
					if idx != -1 {
						vImageTagStr = vImageRepository[:idx]
						vImageRepository = vImageRepository[idx+1:]
					}
				} else {
					vImageTagStr = vImageTag.(string)
				}
				// Last catch all for tag
				if vImageTagStr == "" {
					fmt.Println("WARNING: no version tag found for container. Defaulting to 'Chart.appVersion' for", vImageRepository)
					//vImageTagStr = "latest"
					// default to Chart.AppVersion when this is blank in the values file
					vImageTagStr = chartDef.AppVersion
					// TODO Consider specifying a target version of "sha256-SHAVALUE" when this happens, to avoid CIS Benchmark issues on deployment
				}
				fmt.Println(fmt.Sprintf("- Found container image: '%s/%s:%s@%s'", vImageRegistryStr, vImageRepository, vImageTagStr, vImageDigestStr))
				containers = append(containers, types.ContainerImage{
					Registry:   vImageRegistryStr,
					Repository: vImageRepository,
					Tag:        vImageTagStr,
					Digest:     vImageDigestStr,
				})
				// TODO Save the value mappings of this information so we can override the correct parameters on deployment of the package
			}
		}

		// Now write our summary file
		fmt.Println("Writing kod package file...")
		pkgDef := types.Package{
			Name:       chartDef.Name,
			Version:    chartDef.Version,
			Type:       "helm",
			Sources:    chartDef.Sources,
			Containers: containers,
		}
		pkgBytes, err := yaml.Marshal(pkgDef)
		packageFilePath := filepath.Join(tempPath, "kod-package.yaml")
		err = os.WriteFile(packageFilePath, pkgBytes, os.ModePerm)
		fmt.Println("Written package definition to temporary file", packageFilePath)

		// Copy the Chart folder into a subfolder
		fmt.Println("Copying the helm chart...")
		chartCopyPath := filepath.Join(tempPath, "charts", chartAndVersion)
		// Don't copy if folder already exists
		_, err = os.Stat(chartCopyPath)
		if err == nil {
			fmt.Println("Chart exists, skipping copy of", chartAndVersion)
		} else {
			err = os.MkdirAll(chartCopyPath, os.ModePerm)
			if err != nil {
				fmt.Println("Error copying chart to temporary folder:", chartCopyPath, ",", err)
				os.Exit(1)
			}
			err = os.CopyFS(chartCopyPath, os.DirFS(chartPath))
			if err != nil {
				fmt.Println("Error copying chart to temporary folder from:", chartPath, "to:", chartCopyPath, ",", err)
				os.Exit(1)
			}
		}

		// Copy container images using Skopeo, unless they already exist
		fmt.Println("Fetching any container images required...")
		// TODO Check somehow whether it's a DockerV2 or OCI image repo, and run the appropriate command for this
		ctrFolder := filepath.Join(tempPath, "containers")
		err = os.MkdirAll(ctrFolder, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating containers folder:", ctrFolder, "error:", err)
			os.Exit(1)
		}
		for _, ctr := range containers {
			cf := filepath.Join(ctrFolder, ctr.Registry)
			ctrFile := filepath.Join(cf, ctr.Repository, ctr.Tag+".tar") // changed so that last filename is the tag version, incase container name and version both have hyphens!
			_, err = os.Stat(ctrFile)
			if err == nil {
				fmt.Println("Container folder exists, skipping skopeo copy to", cf)
			} else {
				parent := filepath.Dir(ctrFile)
				err = os.MkdirAll(parent, os.ModePerm)
				if err != nil {
					fmt.Println("Error creating container folder:", cf, "error:", err)
					os.Exit(1)
				}
				// Now invoke skopeo - skopeo copy docker://quay.io/buildah/stable docker-archive:///tmp/kod-redis-1.2.3/containers/docker.io/redis/1.2.3
				srcPath := "docker://" + ctr.Registry + "/" + ctr.Repository + ":" + ctr.Tag
				if ctr.Digest != "" {
					srcPath = "docker://" + ctr.Registry + "/" + ctr.Repository + "@" + ctr.Digest
				}
				skopeoExec := exec.Command("skopeo", "copy", srcPath, "docker-archive:"+ctrFile)
				fmt.Println("Executing", skopeoExec.String())
				err = skopeoExec.Run()
				if err != nil {
					fmt.Println("Error running skopeo copy:", err)
					fmt.Println("Trying backup approach of using tag 'latest' (Needed for many Bitnami container images)")
					srcPath = "docker://" + ctr.Registry + "/" + ctr.Repository + ":latest"
					skopeoExec = exec.Command("skopeo", "copy", srcPath, "docker-archive:"+ctrFile)
					fmt.Println("Executing", skopeoExec.String())
					err = skopeoExec.Run()
					if err != nil {
						os.Exit(1)
					}
				}
				// TODO verify sha, if it exists and we didn't download using the SHA itself
			}
			// TODO if there's an error, delete the folder so it downloads on the next execution
		}

		// Now package the temp folder as a tar.xz but with the kodpkg extension
		files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
			tempPath: folderName,
		})
		if err != nil {
			fmt.Println("Error specifying kod package archive", err)
			os.Exit(1)
		}

		// create the output file we'll write to
		packagePath := filepath.Join(os.TempDir(), folderName+".kodpkg")
		out, err := os.Create(packagePath)
		if err != nil {
			fmt.Println("Error creating kod package archive", err)
			os.Exit(1)
		}
		defer out.Close()

		// we can use the CompressedArchive type to gzip a tarball
		// (since we're writing, we only set Archival, but if you're
		// going to read, set Extraction)
		format := archives.CompressedArchive{
			Compression: archives.Xz{},
			Archival:    archives.Tar{},
		}

		// create the archive
		err = format.Archive(context.Background(), out, files)
		if err != nil {
			fmt.Println("Error writing kod package archive", err)
			os.Exit(1)
		}
		fmt.Println("Package written to", packagePath)
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// packageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// packageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	packageCmd.Flags().StringVarP(&chartPath, "chart", "c", ".", "Folder path containing the helm chart to package")
}
