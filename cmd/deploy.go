/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"kod/internal"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
		err = filepath.WalkDir(ctrFolder, func(path string, d fs.DirEntry, err error) error {
			//fmt.Println("Walking:", path, "Dir?", d.IsDir(), "name:", d.Name())
			if err != nil {
				return err
			}
			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			if !d.IsDir() {
				// Execute skopeo command
				rel, err := filepath.Rel(ctrFolder, path)
				if err != nil {
					fmt.Println("Error finding relative path", path, err)
					os.Exit(1)
				}
				fmt.Println(" - Found container archive", rel)

				parent := filepath.Dir(rel)
				archiveName := d.Name()
				ctrVersion := archiveName
				idx := strings.LastIndex(archiveName, ".")
				if idx != -1 {
					ctrVersion = archiveName[:idx]
				}
				fmt.Println("   - Processing container with repository", parent, "and tag", ctrVersion)

				// Create command and output to the terminal
				destPath := registryUrl
				if strings.HasPrefix(destPath, "http://") {
					destPath = destPath[7:]
				}
				if strings.HasPrefix(destPath, "https://") {
					destPath = destPath[8:]
				}
				if strings.HasPrefix(destPath, "oci://") {
					destPath = destPath[6:]
				}
				if !strings.HasSuffix(destPath, "/") {
					destPath += "/"
				}
				if strings.HasPrefix(projectFolder, "/") {
					destPath += projectFolder[1:]
				} else {
					destPath += projectFolder
				}
				if !strings.HasSuffix(destPath, "/") {
					destPath += "/"
				}
				destPath += parent
				destPath += ":" + ctrVersion

				// TODO try OCI first then docker
				skopeoExec := exec.Command("skopeo", "copy", "docker-archive:"+path, "docker://"+destPath)
				fmt.Println("Executing", skopeoExec.String())

				// Execute the command
				err = skopeoExec.Run()
				if err != nil {
					fmt.Println("Error running skopeo copy. Try skopeo login", registryUrl, "first?", err)
					os.Exit(1)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error walking containers folder path", err)
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

		// TODO perform actual helm install
		fmt.Println("Executing helm...")

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
