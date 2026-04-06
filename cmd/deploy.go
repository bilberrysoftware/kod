/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"fmt"
	"kod/internal"
	"kod/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var skipInstall = false
var cleanup = false
var insecureNoVerify = false
var deploymentName = ""
var targetNamespace = ""
var registryUrl = ""
var projectFolder = ""
var valuesFiles []string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys a kod package to Kubernetes",
	Long: `Unpacks and deploys a kod package to Kubernetes.

First uses skopeo to install all containers into your local container registry.
Then perform a helm upgrade --install on the package.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")

		// TODO sanity check parameter values
		//if projectFolder == "" {
		//	fmt.Println("WARNING projectFolder (-j flag) not set. Defaulting to 'kod/containers'")
		//	projectFolder = "kod/containers"
		//}

		// unpack archive to a temporary folder
		fmt.Println("Unpacking kod package...")
		tmpFolder, err := internal.DoUnpack(context.Background(), packagePath, false)
		if err != nil {
			fmt.Println("Error performing unpack", err)
			os.Exit(1)
		}

		fmt.Println("Reading kod-package.yaml")
		kodPackage := types.Package{}
		packageBytes, err := os.ReadFile(filepath.Join(tmpFolder, "kod-package.yaml"))
		if err != nil {
			fmt.Println("Error reading kod-package.yaml", err)
			os.Exit(1)
		}
		err = yaml.Unmarshal(packageBytes, &kodPackage)
		if err != nil {
			fmt.Println("Error parsing kod-package.yaml", err)
		}

		// load hints file
		hints := types.HintsFile{}
		hintsBytes, err := os.ReadFile(filepath.Join(tmpFolder, "charts", kodPackage.Name+"-"+kodPackage.Version+"-hints.yaml"))
		if err != nil {
			fmt.Println("Error reading hints file", err)
			os.Exit(1)
		}
		err = yaml.Unmarshal(hintsBytes, &hints)
		if err != nil {
			fmt.Println("Error parsing hints file", err)
			os.Exit(1)
		}

		// upload containers to target registry
		fmt.Println("Loading container images into local registry...")
		ctrFolder := filepath.Join(tmpFolder, "containers")
		folder, err := os.Stat(ctrFolder)
		if err != nil {
			fmt.Println(ctrFolder, "does not exist")
			os.Exit(1)
		}
		if !folder.IsDir() {
			fmt.Println("Containers folder", ctrFolder, "is not a directory")
			os.Exit(1)
		}
		// TODO Need to traverse the whole filesystem as we have folders for ctr registry and subfolder
		//fsys := os.DirFS(ctrFolder)
		//if err != nil {
		//	fmt.Println("Error getting containers folder filesystem", err)
		//	os.Exit(1)
		//}
		fmt.Println("Reading containers within folder", ctrFolder)
		// Instead of walking the container folder, read the data from the kod package file
		for _, ctr := range kodPackage.Containers {

			//err = filepath.WalkDir(ctrFolder, func(path string, d fs.DirEntry, err error) error {
			//fmt.Println("Walking:", path, "Dir?", d.IsDir(), "name:", d.Name())
			//if err != nil {
			//	return err
			//}
			//if strings.HasPrefix(d.Name(), ".") {
			//	return fs.SkipDir
			//}
			//if !d.IsDir() {
			paths := []string{tmpFolder}
			paths = append(paths, "containers")
			for _, s := range strings.Split(ctr.Registry, "/") {
				paths = append(paths, s)
			}
			for _, s := range strings.Split(ctr.Repository, "/") {
				paths = append(paths, s)
			}
			ctrFolder = filepath.Join(paths...)
			//// Execute skopeo command
			//rel, err := filepath.Rel(ctrFolder, path)
			//if err != nil {
			//	fmt.Println("Error finding relative path", path, err)
			//	os.Exit(1)
			//}
			//fmt.Println(" - Found container archive", rel)
			//
			//parent := filepath.Dir(rel)
			//archiveName := d.Name()
			//archiveName :=
			//ctrVersion := archiveName
			//idx := strings.LastIndex(archiveName, ".")
			//if idx != -1 {
			//	ctrVersion = archiveName[:idx]
			//}

			fmt.Println("   - Processing container with repository", ctr.Repository, "and tag", ctr.Tag)

			// Create command and output to the terminal
			regPath := registryUrl
			if strings.HasPrefix(regPath, "http://") {
				regPath = regPath[7:]
			}
			if strings.HasPrefix(regPath, "https://") {
				regPath = regPath[8:]
			}
			if strings.HasPrefix(regPath, "oci://") {
				regPath = regPath[6:]
			}
			if !strings.HasSuffix(regPath, "/") {
				regPath += "/"
			}
			repoPath := ""
			if strings.HasPrefix(projectFolder, "/") {
				repoPath += projectFolder[1:]
			} else {
				repoPath += projectFolder
			}
			if ctr.Registry != "" {
				repoPath += "/" + ctr.Registry
			}
			if !strings.HasSuffix(repoPath, "/") {
				repoPath += "/"
			}
			repoPath += ctr.Repository

			destPath := regPath + repoPath

			finalPath := destPath + ":" + ctr.Tag

			// find hint that matches this container image, and change accordingly
			for hIdx, hint := range hints.Hints {
				if hint.PackagedImage.Registry == ctr.Registry &&
					hint.PackagedImage.Tag == ctr.Tag &&
					hint.PackagedImage.Repository == ctr.Repository {
					// replace with our registry and our destPath
					hint.PackagedImage.Registry = regPath
					hint.PackagedImage.Repository = repoPath
					hint.PackagedImage.Tag = ctr.Tag
					if ctr.Digest != "" {
						hint.PackagedImage.Digest = ctr.Digest
					}
					// Replace value in hints
					hints.Hints[hIdx] = hint
				}
			}

			ctrPath := filepath.Join(ctrFolder, ctr.Tag+".tar")

			// TODO try OCI first then docker
			skopeoExec := exec.Command("skopeo", "copy", "docker-archive:"+ctrPath, "docker://"+finalPath)
			fmt.Println("Executing", skopeoExec.String())

			// Execute the command
			err = skopeoExec.Run()
			if err != nil {
				fmt.Println("Error running skopeo copy. Try skopeo login", registryUrl, "first?", err)
				os.Exit(1)
			}
		}
		//	return nil
		//}, )
		//if err != nil {
		//	fmt.Println("Error walking containers folder path", err)
		//	os.Exit(1)
		//}

		// Generate charts/CHARTNAME-CHARTVER-values.yaml file from hints file and any -f inputs
		// generate values file content
		// Note: Any overrides to the values for the local container registry are done in the above code, not here
		valuesFile := map[string]interface{}{}
		for _, hint := range hints.Hints {
			// Create top level structure
			parts := strings.Split(hint.ParentPath, ".")
			lastLevel := valuesFile
			for pIdx, part := range parts {
				if pIdx == len(parts)-1 {
					// Write contents below this
					content := map[string]string{}
					content[hint.RepositoryPath] = hint.PackagedImage.Repository
					if hint.RegistryPath == hint.RepositoryPath+"./" {
						if strings.HasSuffix(hint.PackagedImage.Registry, "/") {
							content[hint.RepositoryPath] = hint.PackagedImage.Registry + content[hint.RepositoryPath]
						} else {
							content[hint.RepositoryPath] = hint.PackagedImage.Registry + "/" + content[hint.RepositoryPath]
						}
					} else {
						content[hint.RegistryPath] = hint.PackagedImage.Registry
					}
					if hint.TagPath == hint.RepositoryPath+".:" {
						content[hint.RepositoryPath] = content[hint.RepositoryPath] + ":" + hint.PackagedImage.Tag
					} else {
						content[hint.TagPath] = hint.PackagedImage.Tag
					}
					if hint.DigestPath != "" {
						if hint.DigestPath == hint.RepositoryPath+".@" {
							content[hint.RepositoryPath] = content[hint.RepositoryPath] + "@" + hint.PackagedImage.Digest
						} else {
							content[hint.DigestPath] = hint.PackagedImage.Digest
						}
					}
					lastLevel[part] = content
				} else {
					newIface := map[string]interface{}{}
					lastLevel[part] = newIface
					lastLevel = newIface
				}
			}
		}
		// write values file
		valuesPath := filepath.Join(tmpFolder, "charts", kodPackage.Name+"-"+kodPackage.Version+"-values.yaml")
		valuesBytes, err := yaml.Marshal(valuesFile)
		if err != nil {
			fmt.Println("Error marshalling values file", err)
			os.Exit(1)
		}
		err = os.WriteFile(valuesPath, valuesBytes, 0644)
		if err != nil {
			fmt.Println("Error writing values file", err)
			os.Exit(1)
		}

		// Conditionally perform helm install
		if skipInstall {
			fmt.Println("Skipping final helm chart deployment due to command line flag --no-install")
			if cleanup {
				fmt.Println("Cleaning up temporary folder", tmpFolder)
				err = os.RemoveAll(tmpFolder)
				if err != nil {
					fmt.Println("Error deleting temporary folder for deployment. Folder:", tmpFolder, "Error:", err)
					os.Exit(1)
				}
			}
			fmt.Println("Done.")
			os.Exit(0)
		}

		// perform actual helm install
		fmt.Println("Executing helm...")
		chartPath := filepath.Join(tmpFolder, "charts", kodPackage.Name+"-"+kodPackage.Version)
		// TODO -f input overrides from command line appended too after our values file
		helmExec := exec.Command("helm", "upgrade", "--install", deploymentName, chartPath, "-n", targetNamespace, "--create-namespace", "-f", valuesPath)
		fmt.Println("Executing", helmExec.String())

		// Execute the command
		err = helmExec.Run()
		if err != nil {
			fmt.Println("Error running helm upgrade --install.", err)
			os.Exit(1)
		}

		if cleanup {
			fmt.Println("Cleaning up temporary folder", tmpFolder)
			err = os.RemoveAll(tmpFolder)
			if err != nil {
				fmt.Println("Error deleting temporary folder for deployment. Folder:", tmpFolder, "Error:", err)
				os.Exit(1)
			}
		}
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deployCmd.Flags().BoolVar(&skipInstall, "no-install", false, "Performs everything but the final helm upgrade --install command")
	deployCmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove temporary folder after successful command execution")
	deployCmd.Flags().BoolVar(&insecureNoVerify, "insecure-no-verify", false, "Do not verify server TLS certs in the Registry or Kubernetes")
	deployCmd.Flags().StringVarP(&packagePath, "package", "p", "", ".kodpkg file to deploy")
	deployCmd.Flags().StringVarP(&deploymentName, "deployment", "d", "", "Deployment name for helm")
	deployCmd.Flags().StringVarP(&targetNamespace, "namespace", "n", "default", "Target Kubernetes Namespace")
	deployCmd.Flags().StringVarP(&registryUrl, "registry", "r", "", "The base URL for the container registry to use")
	deployCmd.Flags().StringVarP(&projectFolder, "project", "j", "kod/containers", "The project path within the container registry to use as the base folder")
	valuesFiles = *deployCmd.Flags().StringArrayP("values", "f", []string{}, "The helm values file(s) to use to customise a deployment")

	// TODO oci/helm artifact project path of -a (defaults to kod/charts
	// TODO oci registry URL --oci (defaults to same as registryUrl)
	// TODO --ci flag to not prompt (if applicable)
	// TODO --output-commands-only flag to only output commands to STDOUT without execution, or any other logging (suppresses --cleanup)
}
