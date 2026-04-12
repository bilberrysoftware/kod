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
var helmWait = false
var deploymentName = ""
var targetNamespace = ""
var registryUrl = ""
var projectFolder = ""
var valuesFiles []string
var additionalImagePullSecrets []string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys a kod package to Kubernetes",
	Long: `Unpacks and deploys a kod package to Kubernetes.

First uses skopeo to install all containers into your local container registry.
Then perform a helm upgrade --install on the package.

Examples:
  kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://myregistry:8080 -d mycnpg -s zot                 # Unpack, copy to registry, deploy to the default namespace, and use existing registry secret called 'zot'
  kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://myregistry:8080 -d mycnpg -s zot -j public_copy  # As above, but use the /public_copy/ project folder for containers instead of /kod/containers/
  kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://myregistry:8080 -n mynamespace -d mycnpg         # Unpack, copy to registry, deploy to mynamespace
  kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://myregistry:8080 --no-install -d mycnpg           # Unpack, copy to registry only (populates helm values files with new container locations)
`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("deploy called")

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
		//if projectFolder == "" {
		//	fmt.Println("WARNING projectFolder (-j flag) not set. Defaulting to 'kod/containers'")
		//	projectFolder = "kod/containers"
		//}

		if strings.HasSuffix(registryUrl, "/") {
			registryUrl = registryUrl[:len(registryUrl)-1]
		}
		// Check that skopeo is logged in already
		skopeoLoginExec := exec.Command("skopeo", "login", "--get-login", registryUrl)

		var skopeoLoginOutput internal.SaveOutput
		skopeoLoginOutput.NoEchoToStdOut = true
		skopeoLoginExec.Stdin = os.Stdin
		skopeoLoginExec.Stdout = &skopeoLoginOutput
		skopeoLoginExec.Stderr = os.Stderr

		err := skopeoLoginExec.Run()
		if err != nil {
			fmt.Println("Error executing skopeo login check against target registry:", err, "details:", skopeoLoginOutput.String())
			fmt.Println("You must be logged in to your container registry to be able to upload container images. Exiting.")
			os.Exit(1)
		}
		fmt.Println("Skopeo login successful as user:", strings.TrimSpace(skopeoLoginOutput.String()), "to registry:", registryUrl)

		// TODO Check that helm/kubectl is logged into target cluster just prior to deployment (ignore if --no-install is set)

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

		containerFileMap := make(map[string]string)
		fallbackTagPaths := make(map[string]string)

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

			fmt.Println("Processing container with repository", ctr.Repository, "and tag", ctr.Tag)

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
			//if !strings.HasSuffix(regPath, "/") {
			//	regPath += "/"
			//}
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

			destPath := regPath + "/" + repoPath

			finalPath := destPath + ":" + ctr.Tag

			// find hint that matches this container image, and change accordingly
			// do for all subchart hints too (these are in the top level HintsFile by this point)
			for hIdx, hint := range hints.Images {
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
					hints.Images[hIdx] = hint
					continue
				}
				// Check to see if we have a version mismatch
				if hint.PackagedImage.Registry == ctr.Registry &&
					hint.PackagedImage.Repository == ctr.Repository {
					fmt.Println("")
					if !internal.ContainerVersionExists(kodPackage.Containers, hint.PackagedImage) {
						fmt.Println("WARNING Hint has an invalid container tag. Setting to this available container tag. Reg:", hint.PackagedImage.Registry, "Repo:", hint.PackagedImage.Repository, "Tag:", hint.PackagedImage.Tag, "->", ctr.Tag)

						// replace with our registry and our destPath
						hint.PackagedImage.Registry = regPath
						hint.PackagedImage.Repository = repoPath
						hint.PackagedImage.Tag = ctr.Tag
						if ctr.Digest != "" {
							hint.PackagedImage.Digest = ctr.Digest
						}
						// Replace value in hints
						hints.Images[hIdx] = hint
						continue
					}
				}
			}

			ctrPath := filepath.Join(ctrFolder, ctr.Tag+".tar")

			_, err := os.Stat(ctrPath)
			if err != nil {
				// File doesn't exist
				fmt.Println("WARNING: Container archive doesn't exist. Skipping copy of this archive. Archive path:", ctrPath)
			} else {
				//containerFileMap[ctr.Registry+"/"+ctr.Repository+":"+ctr.Tag] = ctrPath
				//fallbackTagPaths[ctr.Registry+"/"+ctr.Repository] = ctr.Tag
				containerFileMap[regPath+"/"+repoPath+":"+ctr.Tag] = ctrPath
				fallbackTagPaths[regPath+"/"+repoPath] = ctr.Tag

				// TODO try OCI first then docker
				skopeoExec := exec.Command("skopeo", "copy", "docker-archive:"+ctrPath, "docker://"+finalPath, "--retry-times", "5")
				fmt.Println("Executing", skopeoExec.String())
				var skopeoOutput internal.SaveOutput
				skopeoOutput.Prefix = "  \xF0\x9F\x9A\xA2 "
				skopeoExec.Stdin = os.Stdin
				skopeoExec.Stdout = &skopeoOutput
				skopeoExec.Stderr = os.Stderr

				// Execute the command
				err := skopeoExec.Run()
				if err != nil {
					fmt.Println("Error running skopeo copy. Try skopeo login", registryUrl, "first?", err, "details:", skopeoOutput.String())
					os.Exit(1)
				}
			}
		}
		//	return nil
		//}, )
		//if err != nil {
		//	fmt.Println("Error walking containers folder path", err)
		//	os.Exit(1)
		//}

		fmt.Println("Container map:-")
		for ctrTag, ctrPath := range containerFileMap {
			fmt.Println(ctrTag, "=", ctrPath)
		}
		fmt.Println("Container fallback:-")
		for ctrRepo, fallbackVersion := range fallbackTagPaths {
			fmt.Println(ctrRepo, "=", fallbackVersion)
		}

		// TODO do the below for ALL charts, not just the main one, so container refs are correct
		// TODO do we need to do this in the main chart only, with values pointing to lower charts?
		//  - YES. See https://helm.sh/docs/chart_template_guide/subcharts_and_globals/#overriding-values-from-a-parent-chart

		// Generate charts/CHARTNAME-CHARTVER-values.yaml file from hints file and any -f inputs
		// generate values file content
		// Note: Any overrides to the values for the local container registry are done in the above code, not here
		valuesFile := map[string]interface{}{}
		for _, hint := range hints.Images {
			// Create top level structure
			parts := strings.Split(hint.ParentPath, ".")
			lastLevel := valuesFile
			for pIdx, part := range parts {
				if pIdx == len(parts)-1 {
					// Write contents below this
					content := map[string]string{}
					content[hint.RepositoryPath] = hint.PackagedImage.Repository
					//pathPrefix := ""
					//if hint.ParentPath != "" {
					//	pathPrefix = hint.ParentPath + "."
					//}
					// Parse out any refs that might be within another values element
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
						// prevent both :tag@digest in final URL, preferring @digest only
						if hint.DigestPath != hint.RepositoryPath+".@" {
							content[hint.RepositoryPath] = content[hint.RepositoryPath] + ":" + hint.PackagedImage.Tag
						}
					} else {
						content[hint.TagPath] = hint.PackagedImage.Tag
					}

					//if -1 != strings.Index(hint.PackagedImage.Repository, "kube-state-metrics") {
					//	fmt.Println("Found kube-state-metrics.")
					//	fmt.Println(" - HINT reg:", hint.PackagedImage.Registry, "rep:", hint.PackagedImage.Repository, "tag:", hint.PackagedImage.Tag)
					//	fmt.Println(" - content reg:", content[hint.RegistryPath], "rep:", content[hint.RepositoryPath], "tag:", content[hint.TagPath])
					//}

					// ensure container at TagPath exists, and if not, fallback to version in main kod-package.yaml file
					//ctrId := content[hint.RegistryPath] + "/" + content[hint.RepositoryPath]
					ctrId := hint.PackagedImage.Registry + "/" + hint.PackagedImage.Repository
					//ctrTarFile := containerFileMap[ctrId+":"+content[hint.TagPath]]
					ctrTarFile := containerFileMap[ctrId+":"+hint.PackagedImage.Tag]
					if ctrTarFile == "" {
						fmt.Println(fmt.Sprintf("WARNING: Version '%s' in helm chart for container '%s' isn't available in archive. Attempting fallback version", hint.PackagedImage.Tag, ctrId))
						fallback := fallbackTagPaths[ctrId]
						if fallback == "" {
							fmt.Println(" - WARNING: No valid fallback tag value detected for container:", ctrId)
						} else {
							fmt.Println(" - Found valid fallback tag:", fallback)
							content[hint.TagPath] = fallback
						}
					}
					//if -1 != strings.Index(hint.PackagedImage.Repository, "kube-state-metrics") {
					//	fmt.Println(" - content NOW reg:", content[hint.RegistryPath], "rep:", content[hint.RepositoryPath], "tag:", content[hint.TagPath])
					//	//os.Exit(1)
					//}

					// (Note: This usually happens because the value in the helm chart is wrong, or doesn't exist)
					if hint.DigestPath != "" {
						if hint.DigestPath == hint.RepositoryPath+".@" {
							content[hint.RepositoryPath] = content[hint.RepositoryPath] + "@" + hint.PackagedImage.Digest
						} else {
							content[hint.DigestPath] = hint.PackagedImage.Digest
						}
					}
					lastLevel[part] = content
				} else {
					// Otherwise if we have two containers under the same registry, we only declare one image!
					if nil == lastLevel[part] {
						newIface := map[string]interface{}{}
						lastLevel[part] = newIface
						//lastLevel = newIface
					}
					lastLevel = lastLevel[part].(map[string]interface{})
				}
			}
		}

		// Now pass in imagePullSecrets overrides
		for _, ipsh := range hints.ImagePullSecrets {
			if "" != ipsh.SecretArrayPath {
				//fmt.Println("DEBUG: Got additional secrets:", additionalImagePullSecrets)
				//fmt.Println("DEBUG: path:", ipsh.SecretArrayPath)
				parts := strings.Split(ipsh.SecretArrayPath, ".")
				lastLevel := valuesFile
				for pIdx, part := range parts {
					//fmt.Println("DEBUG: loop idx", pIdx, "part", part)
					if pIdx == len(parts)-1 {
						// Write contents below this
						var content []types.SecretReference
						var secretArray []types.SecretReference
						for _, value := range ipsh.PackagedSecrets {
							secretArray = append(secretArray, value)
						}
						for _, value := range additionalImagePullSecrets {
							secretArray = append(secretArray, types.SecretReference{
								Name: value,
							})
						}
						//fmt.Println("DEBUG: secret array", secretArray)
						for _, secret := range secretArray {
							content = append(content, secret)
						}
						lastLevel[part] = content
						//fmt.Println("DEBUG: last level final content", content, "for part", part)
					} else {
						// Note sure this is needed currently, as we don't loop, but just in case I'll leave it here
						if nil == lastLevel[part] {
							newIface := map[string]interface{}{}
							lastLevel[part] = newIface
						}
						lastLevel = lastLevel[part].(map[string]interface{})
						//fmt.Println("DEBUG: intermediate level", part)
					}
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

		// END CHART VALUES GENERATION

		// Conditionally perform helm install
		if skipInstall {
			fmt.Println("Skipping final helm chart deployment due to command line flag --no-install")
			if cleanup {
				fmt.Println("Cleaning up temporary folder", tmpFolder)
				err = os.RemoveAll(tmpFolder)
				if err != nil {
					fmt.Println("WARNING: Error deleting temporary folder for deployment. Folder:", tmpFolder, "Error:", err)
					//os.Exit(1)
				}
			}
			fmt.Println("Done.")
			os.Exit(0)
		}

		// perform actual helm install
		fmt.Println("Executing helm...")
		chartPath := filepath.Join(tmpFolder, "charts", kodPackage.Name+"-"+kodPackage.Version)

		cmdArgs := []string{"upgrade", "--install", deploymentName, chartPath, "-n", targetNamespace, "--create-namespace", "-f", valuesPath}
		// Include -f values file overrides from command line appended too after our values file
		for _, vf := range valuesFiles {
			cmdArgs = append(cmdArgs, "-f")
			cmdArgs = append(cmdArgs, vf)
		}
		if helmWait {
			cmdArgs = append(cmdArgs, "--wait")
		}
		helmExec := exec.Command("helm", cmdArgs...)
		fmt.Println("Executing", helmExec.String())
		var helmOutput internal.SaveOutput
		helmOutput.Prefix = "  \xF0\x9F\x8C\x90 "
		helmExec.Stdin = os.Stdin
		helmExec.Stdout = &helmOutput
		helmExec.Stderr = os.Stderr

		// Execute the command
		err = helmExec.Run()
		// Execute the command
		if err != nil {
			fmt.Println("Error running helm upgrade --install.", err, "details:", helmOutput.String())
			os.Exit(1)
		}

		if cleanup {
			fmt.Println("Cleaning up temporary folder", tmpFolder)
			err = os.RemoveAll(tmpFolder)
			if err != nil {
				fmt.Println("WARNING: Error deleting temporary folder for deployment. Folder:", tmpFolder, "Error:", err)
				//os.Exit(1)
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
	deployCmd.Flags().BoolVar(&helmWait, "wait", false, "Pass the --wait parameter to the helm upgrade --install command")
	deployCmd.Flags().StringVarP(&packagePath, "package", "p", "", ".kodpkg file to deploy")
	deployCmd.Flags().StringVarP(&deploymentName, "deployment", "d", "", "Deployment name for helm")
	deployCmd.Flags().StringVarP(&targetNamespace, "namespace", "n", "default", "Target Kubernetes Namespace")
	deployCmd.Flags().StringVarP(&registryUrl, "registry", "r", "", "The base URL for the container registry to use")
	deployCmd.Flags().StringVarP(&projectFolder, "project", "j", "kod/containers", "The project path within the container registry to use as the base folder")
	deployCmd.Flags().StringArrayVarP(&valuesFiles, "values", "f", []string{}, "The helm values file(s) to use to customise a deployment")
	deployCmd.Flags().StringArrayVarP(&additionalImagePullSecrets, "imagePullSecrets", "s", []string{}, "Any additional imagePullSecrets in the target namespace to use")

	// TODO oci/helm artifact project path of -a (defaults to kod/charts
	// TODO oci registry URL --oci (defaults to same as registryUrl)
	// TODO --ci flag to not prompt (if applicable)
	// TODO --output-commands-only flag to only output commands to STDOUT without execution, or any other logging (suppresses --cleanup)
}
