/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"kod/internal"
	"kod/types"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mholt/archives"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var chartPath string
var chartRegistry string

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Packages an individual Helm chart or helmfile",
	Long: `Creates a deployable package archive consistent of the helm chart or helm file
and all of its constituent container images.`,
	Run: func(cmd *cobra.Command, args []string) {

		// sanity check that our dependent commands exist
		if !internal.CommandExists("skopeo") {
			fmt.Println("skopeo not installed. Cannot proceed. Exiting.")
			os.Exit(1)
		}
		if !internal.CommandExists("helm") {
			fmt.Println("helm not installed. Cannot proceed. Exiting.")
			os.Exit(1)
		}

		// TODO sanity check parameter values

		// Default chartFolder to chartPath
		chartFolder := chartPath

		fmt.Println("chart registry:", chartRegistry)
		// If chart is remote, fetch using helm repo add/update and helm save, then unpack the tar (Don't repack it later - just copy and rename)
		if "" != chartRegistry {
			fmt.Println("Using chart registry:", chartRegistry)
			// create helm local repo name from last part of URL
			// Known examples:-
			// - Redis Operator
			//   - helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
			//   - helm install redis-operator ot-helm/redis-operator --namespace ot-operators --set featureGates.GenerateConfigInInitContainer=true
			// - cert-manager
			//   - helm repo add jetstack https://charts.jetstack.io
			//   - helm install cert-manager --namespace cert-manager --version v1.17.2 jetstack/cert-manager
			// - Nifikop Operator (OCI Artifact URL)
			//   - helm install nifikop oci://ghcr.io/konpyutaika/helm-charts/nifikop --namespace=nifi --version 0.13.0 ...

			// Generate temporary file name
			predNameString := chartRegistry + chartPath
			h := sha256.New()
			h.Write([]byte(predNameString))
			predictableNameHash := h.Sum(nil)
			predictableName := base64.RawStdEncoding.EncodeToString(predictableNameHash)
			//tempChartArchive := filepath.Join(os.TempDir(), "kod-"+predictableName+".tgz")
			chartFolder = filepath.Join(os.TempDir(), "kod-"+predictableName)

			//chartName := ""
			//chartVersion := ""

			// Fetch chart
			if strings.HasPrefix(chartRegistry, "https://") {
				// We have a https helm3 registry URL. Name of the chart is in the -c option
				// Now fetch the chart from a HTTPS URL (Public with no auth)

				helmRepoAddExec := exec.Command("helm", "repo", "add", predictableName, chartRegistry, "--force-update")
				fmt.Println("Executing", helmRepoAddExec.String())
				helmRepoAddOutput, err := helmRepoAddExec.Output()
				if err != nil {
					fmt.Println("Error executing helm repo add:", err, "details:", helmRepoAddOutput)
					os.Exit(1)
				}

				helmRepoUpdateExec := exec.Command("helm", "repo", "update")
				fmt.Println("Executing", helmRepoUpdateExec.String())
				helmRepoUpdateOutput, err := helmRepoUpdateExec.Output()
				if err != nil {
					fmt.Println("Error executing helm repo update:", err, "details:", helmRepoUpdateOutput)
					os.Exit(1)
				}

				// use helm search to find our target chart
				//helmRepoSearchExec := exec.Command("helm", "repo", "search", "-oyaml")
				//fmt.Println("Executing", helmRepoSearchExec.String())
				//searchOutput, err := helmRepoSearchExec.Output()
				//if err != nil {
				//	fmt.Println("Error executing helm repo search:", err)
				//	os.Exit(1)
				//}
				//// Now get the output from this command and parse it
				//searchResults := types.HelmSearchResults{}
				//err = yaml.Unmarshal(searchOutput, &searchResults)
				//if err != nil {
				//	fmt.Println("Error unmarshalling helm search results:", err)
				//	os.Exit(1)
				//}
				//// Now find the chart we've got in chartPath, but don't forget our prefix
				//shortRef := predictableName + "/" + chartPath
				//searchResult := types.HelmSearchResult{}
				//found := false
				//for _, sr := range searchResults {
				//	if sr.Name == shortRef {
				//		found = true
				//		searchResult = sr
				//	}
				//}
				//if found {
				//	fmt.Println(fmt.Sprintf("Found Helm chart %s in %s (Registry: %s)", chartPath, shortRef, chartRegistry))
				//	chartName = chartPath
				//	chartVersion = searchResult.Version
				//}

				// Now save the chart to our temporary archive
				helmPullExec := exec.Command("helm", "pull", predictableName+"/"+chartPath, "--untar", "--untardir", chartFolder)
				fmt.Println("Executing", helmPullExec.String())
				helmPullOutput, err := helmPullExec.Output()
				if err != nil {
					fmt.Println("Error executing helm pull:", err, "details:", helmPullOutput)
					os.Exit(1)
				}

				// Note that helm pull uses the chart name as a subfolder
				chartFolder = filepath.Join(chartFolder, chartPath)
				fmt.Println("Saved helm chart to " + chartFolder)

				// Now unarchive it - it's a tar.gz file

				// Use archives to unpack this into the temporary folder (includes path within archive)
				//format := archives.CompressedArchive{
				//	Compression: archives.Gz{},
				//	Extraction:  archives.Tar{},
				//}
				//fh, err := os.Open(tempChartArchive)
				//if err != nil {
				//	fmt.Println("Error opening chart tgz archive file:", err)
				//	os.Exit(1)
				//}
				//defer fh.Close()
				//outAbs, err := filepath.Abs(chartFolder)
				//if err != nil {
				//	fmt.Println(fmt.Errorf("calling filepath.Abs on output dir '%s' failed: %w", chartFolder, err))
				//	os.Exit(1)
				//}
				//err = format.Extract(context.Background(), fh,
				//	func(ctx context.Context, fi archives.FileInfo) error {
				//		nameInArchive := fi.NameInArchive
				//
				//		if nameInArchive == "" || nameInArchive == "." {
				//			return nil
				//		}
				//
				//		cleanName := filepath.Clean(nameInArchive)
				//		destPath := filepath.Join(outAbs, cleanName)
				//
				//		destAbs, err := filepath.Abs(destPath)
				//		if err != nil {
				//			return fmt.Errorf("calling filepath.Abs on dest path '%s' failed: %w", destPath, err)
				//		}
				//
				//		// Avoid traversal attacks
				//		if !strings.HasPrefix(destAbs, outAbs+string(os.PathSeparator)) && destAbs != outAbs {
				//			return fmt.Errorf("unsafe path in archive: %q", nameInArchive)
				//		}
				//
				//		// Create directory if in archive
				//		info, err := fi.Stat()
				//		if err != nil {
				//			return fmt.Errorf("stat on %q failed: %w", nameInArchive, err)
				//		}
				//		if info.IsDir() {
				//			return os.MkdirAll(destAbs, 0o755)
				//		}
				//
				//		// Ensure parent directories exist
				//		if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
				//			return fmt.Errorf("mkdir on parent '%s' failed: %w", destAbs, err)
				//		}
				//
				//		// Open archive entry for reading
				//		rc, err := fi.Open()
				//		if err != nil {
				//			return fmt.Errorf("open entry %q failed: %w", nameInArchive, err)
				//		}
				//		defer rc.Close()
				//
				//		// Create destination file
				//		outFile, err := os.OpenFile(destAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
				//		if err != nil {
				//			return fmt.Errorf("create on %q failed: %w", destAbs, err)
				//		}
				//		defer outFile.Close()
				//
				//		// Copy contents
				//		if _, err := io.Copy(outFile, rc); err != nil {
				//			return fmt.Errorf("copy on %q failed: %w", nameInArchive, err)
				//		}
				//
				//		return nil
				//	})
				//if err != nil {
				//	fmt.Println("Error unpacking chart tgz:", err)
				//	os.Exit(1)
				//}
				fmt.Println(fmt.Sprintf("Chart '%s' from registry '%s' unpacked to temporary folder '%s'", chartPath, chartRegistry, chartFolder))

				//resp, err := http.Get(chartRegistry)
				//if err != nil {
				//	fmt.Println("Error fetching chart at registry:", chartRegistry, err)
				//	os.Exit(1)
				//}
				//defer resp.Body.Close()
				//
				//// TODO handle redirects
				//if resp.StatusCode != http.StatusOK {
				//	fmt.Println("Failed to fetch chart at registry:", chartRegistry, "Error Status not OK:", resp.StatusCode)
				//	os.Exit(1)
				//}
				//// Now stream the response to our file
				//body,err := io.ReadAll(resp.Body)
				//if err != nil {
				//	fmt.Println("Error reading chart registry fetch response body:", err)
				//	os.Exit(1)
				//}
				//err = os.WriteFile(tempChartArchive, body, 0644)
				//if err != nil {
				//	fmt.Println("Error creating temporary chart archive for chart registry:", err)
				//}

			} else {
				if strings.HasPrefix(chartRegistry, "oci://") {
					// We have an OCI registry url, which INCLUDES the -c path. So if -c is included, warn but append to URL
					if "" != chartPath && "." != chartPath {
						newOciUrl := chartRegistry
						if !strings.HasSuffix(newOciUrl, "/") {
							newOciUrl += "/"
						}
						newOciUrl += chartPath
						fmt.Println(fmt.Sprintf("WARNING: You are using an OCI Registry (-r) URL with a chart path (-c). Normally, you just use the full OCI URL with the -r option. We've rewritten your OCI URL to: %s", newOciUrl))
						chartPath = ""
						chartRegistry = newOciUrl
					}

					// TODO unpack chart
					// Try direct download using helm pull with OCI format URL

					helmPullExec := exec.Command("helm", "pull", chartRegistry, "--untar", "--untardir", chartFolder)
					fmt.Println("Executing", helmPullExec.String())
					helmPullOutput, err := helmPullExec.Output()
					if err != nil {
						fmt.Println("Error executing OCI helm pull:", err, "details:", helmPullOutput)
						os.Exit(1)
					}

					// Note that helm pull uses the chart name as a subfolder
					// Take the last part of the OCI URL as the chart folder name
					idx := strings.LastIndex(chartRegistry, "/")
					if -1 == idx {
						fmt.Println("Error extracting chart name from OCI URL:", chartRegistry, err)
						os.Exit(1)
					}
					chartPath = chartRegistry[idx+1:]
					chartFolder = filepath.Join(chartFolder, chartPath)
					fmt.Println("Saved helm chart to " + chartFolder)

				} else {
					fmt.Println("Unsupported helm registry URL scheme (We support https:// and oci://):", chartRegistry)
				}
			}

			// Unpack to a temporary folder
			// set chartFolder appropriately
		} else {
			fmt.Println("Using local chart folder:", chartPath)
		}

		// See if the Chart.yaml file exists, error if not
		folder, err := os.Stat(chartFolder)
		if err != nil {
			fmt.Println(chartFolder, "does not exist")
			os.Exit(1)
		}
		if !folder.IsDir() {
			fmt.Println("Chart folder", chartFolder, "is not a directory")
			os.Exit(1)
		}
		chartYaml := filepath.Join(chartFolder, "Chart.yaml")
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
		fmt.Println(fmt.Sprintf("Packaging chart. name: '%s', version: '%s', appVersion: `%s`",
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
		valuesPath := filepath.Join(chartFolder, "values.yaml")
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

		// Create hints file for all that we learn about the container images used by this chart
		hints := types.HintsFile{
			ChartRef: types.HelmChartReference{
				Name:    chartDef.Name,
				Version: chartDef.Version,
			},
			Images: []types.ContainerImageHint{},
		}

		// Now look for various common image config locations
		// Option 1. Helm create default single container values.yaml elements
		vImageEl := valuesMap["image"]
		if vImageEl != nil {
			vImage := vImageEl.(map[string]interface{})
			err = internal.FindContainerImagesByImageChildValues(chartDef, "image", &vImage, &containers, &hints)
			if err != nil {
				fmt.Println("Error finding image child elements in values.yaml", err)
				os.Exit(1)
			}
		}
		// End Option 1.

		// Option 2. Underneath any element in values.yaml with a parent called .*[iI]mage:
		err = internal.FindContainerImagesByImageTagSearch(chartDef, "", &valuesMap, &containers, &hints)
		if err != nil {
			fmt.Println("Error finding container images by depth first values.yaml search", err)
			os.Exit(1)
		}
		// End Option 2.

		// Now search for ImagePullSecrets
		fmt.Println("Searching for imagePullSecrets...")
		globalEl := valuesMap["global"]
		if globalEl != nil {
			// TODO ensure type is map before case
			global := globalEl.(map[string]interface{})
			ipsEl := global["imagePullSecrets"]
			if ipsEl != nil {
				// This is if it has a value specified. It will be an empty interface{} if blank (which is the norm)
				if reflect.TypeOf(ipsEl) == reflect.TypeOf([]string{}) {
					ips := ipsEl.([]string)
					for _, secretRef := range ips {
						hints.ImagePullSecrets.PackagedSecrets = append(hints.ImagePullSecrets.PackagedSecrets, types.SecretReference{Name: secretRef})
					}
					//} else {
					//	hints.ImagePullSecrets.PackagedSecrets = []string{}
				}
				hints.ImagePullSecrets.SecretArrayPath = "global.imagePullSecrets"
			}
		} else {
			hints.ImagePullSecrets.PackagedSecrets = []types.SecretReference{}
			ipsEl := valuesMap["imagePullSecrets"]
			if ipsEl != nil {
				if reflect.TypeOf(ipsEl) == reflect.TypeOf([]string{}) {
					ips := ipsEl.([]string)
					for _, secretRef := range ips {
						hints.ImagePullSecrets.PackagedSecrets = append(hints.ImagePullSecrets.PackagedSecrets, types.SecretReference{Name: secretRef})
					}

					//} else {
					//	nameMap := map[string]interface{}{}
					//	hints.ImagePullSecrets.PackagedSecrets = []string{}
				}
				hints.ImagePullSecrets.SecretArrayPath = "imagePullSecrets"
			}
		}

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
			err = os.CopyFS(chartCopyPath, os.DirFS(chartFolder))
			if err != nil {
				fmt.Println("Error copying chart to temporary folder from:", chartFolder, "to:", chartCopyPath, ",", err)
				os.Exit(1)
			}
			// Copy hints file over
			fmt.Println(" - Writing hints file for chart")
			hintsFileName := filepath.Join(tempPath, "charts", chartAndVersion+"-hints.yaml")
			hintsBytes, err := yaml.Marshal(hints)
			if err != nil {
				fmt.Println("Error marshaling hints file", err)
				os.Exit(1)
			}
			err = os.WriteFile(hintsFileName, hintsBytes, os.ModePerm)
			if err != nil {
				fmt.Println("Error writing hints file", err)
				os.Exit(1)
			}
		}

		// TODO ensure we only mention each container image once in containers
		// Some things like Kafka, kube-prometheus-stack use the same container image with different runtime settings, and refer to it multiple times in the same chart

		// Copy container images using Skopeo, unless they already exist
		fmt.Println("Fetching any container images required...")
		// TODO Check somehow whether it's a DockerV2 or OCI image repo, and run the appropriate command for this
		ctrFolder := filepath.Join(tempPath, "containers")
		err = os.MkdirAll(ctrFolder, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating containers folder:", ctrFolder, "error:", err)
			os.Exit(1)
		}
		for ctrIdx, ctr := range containers {
			cf := filepath.Join(ctrFolder, ctr.Registry)
			ctrFile := filepath.Join(cf, ctr.Repository, ctr.Tag+".tar") // changed so that last filename is the tag version, incase container name and version both have hyphens!
			_, err = os.Stat(ctrFile)
			if err == nil {
				fmt.Println("Container folder exists, skipping skopeo copy to", ctrFile)
			} else {
				parent := filepath.Dir(ctrFile)
				err = os.MkdirAll(parent, os.ModePerm)
				if err != nil {
					fmt.Println("Error creating container folder:", ctrFile, "error:", err)
					os.Exit(1)
				}

				// try to inspect the container image now to list available tags
				// WARNING: Not specifying a version actually looks for a 'latest' tag, you MUST specify a version if known
				imagePath := "docker://" + ctr.Registry + "/" + ctr.Repository
				if ctr.Tag != "" {
					imagePath += ":" + ctr.Tag
				}
				skopeoInspectExec := exec.Command("skopeo", "inspect", imagePath)
				fmt.Println("Executing", skopeoInspectExec.String())
				var skopeoInspectOutput internal.SaveOutput
				skopeoInspectOutput.NoEchoToStdOut = true
				skopeoInspectExec.Stdin = os.Stdin
				skopeoInspectExec.Stdout = &skopeoInspectOutput
				skopeoInspectExec.Stderr = os.Stderr
				err = skopeoInspectExec.Run()
				if err != nil {
					fmt.Println("Error inspecting skopeo image:", ctrFile, "error:", err, "details:", skopeoInspectOutput.String())
					os.Exit(1)
				}
				// Now parse output and set Digest to this value
				var inspectData = types.SkopeoInspectResult{}
				err = yaml.Unmarshal(skopeoInspectOutput.SavedOutput, &inspectData)
				if err != nil {
					fmt.Println("Error unmarshaling inspect result:", ctrFile, "error:", err)
					os.Exit(1)
				}
				// If Digest already set, validate the digest is the SAME
				if "" != ctr.Digest && "" != inspectData.Digest {
					if inspectData.Digest != ctr.Digest {
						fmt.Println("WARNING: Declared digest for container in helm and in registry are different. Helm digest:", ctr.Digest, "registry digest:", inspectData.Digest, "Continuing anyway. Overwriting Digest value.")
					}
				}
				// Set container digest
				ctr.Digest = inspectData.Digest
				containers[ctrIdx] = ctr // sets the value

				// TODO change the below to verify if the requested tag is available
				// Now invoke skopeo - skopeo copy docker://quay.io/buildah/stable docker-archive:///tmp/kod-redis-1.2.3/containers/docker.io/redis/1.2.3
				srcPath := "docker://" + ctr.Registry + "/" + ctr.Repository + ":" + ctr.Tag
				if ctr.Digest != "" {
					srcPath = "docker://" + ctr.Registry + "/" + ctr.Repository + "@" + ctr.Digest
				}
				skopeoExec := exec.Command("skopeo", "copy", srcPath, "docker-archive:"+ctrFile, "--retry-times", "5")
				var skopeoOutput internal.SaveOutput
				skopeoExec.Stdin = os.Stdin
				skopeoExec.Stdout = &skopeoOutput
				skopeoExec.Stderr = os.Stderr
				fmt.Println("Executing", skopeoExec.String())
				err := skopeoExec.Run()
				if err != nil {
					fmt.Println("Error running skopeo copy:", err, "details:", skopeoOutput.String())
					fmt.Println("Trying backup approach of using tag 'latest' (Needed for many Bitnami container images)")
					srcPath = "docker://" + ctr.Registry + "/" + ctr.Repository + ":latest"
					skopeoOutput.Clear()
					skopeoExec = exec.Command("skopeo", "copy", srcPath, "docker-archive:"+ctrFile, "--retry-times", "5")
					skopeoExec.Stdin = os.Stdin
					skopeoExec.Stdout = &skopeoOutput
					skopeoExec.Stderr = os.Stderr
					fmt.Println("Executing", skopeoExec.String())
					err = skopeoExec.Run()
					if err != nil {
						fmt.Println("Error running skopeo copy using latest tag:", err, "details:", skopeoOutput.String())
						os.Exit(1)
					}
				}
				// TODO verify sha, if it exists and we didn't download using the SHA itself
			}
			// TODO if there's an error, delete the folder so it downloads on the next execution
		}

		// Now write our summary file - done here so container digests are known
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

		fmt.Println("Creating final package archive (.kodpkg) ...")
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
	packageCmd.Flags().StringVarP(&chartPath, "chart", "c", ".", "Folder path containing the local helm chart to package, or the repository path if a remote chart (See -r)")
	packageCmd.Flags().StringVarP(&chartRegistry, "registry", "r", "", "Registry URL of helm or OCI repo if the chart is remote")
}
