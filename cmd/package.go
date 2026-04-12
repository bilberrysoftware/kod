/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"kod/internal"
	"kod/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
and all of its constituent container images.

Examples:
  kod package -c ./charts/my-local-chart-folder                                     # Packages from a local unzipped Helm chart folder
  kod package -c ./charts/my-local-chart.tgz                                        # Packages from a local Helm chart archive tgz file
  kod package -r https://ot-container-kit.github.io/helm-charts/ -c redis-operator  # Packages from a remote HTTPS based Helm chart repository
  kod package -r oci://ghcr.io/konpyutaika/helm-charts -c nifikop                   # Packages from a remote OCI based registry
`,
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

		warnings := []types.ComplianceWarning{}

		//fmt.Println("chart registry:", chartRegistry)
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
			predictableName := hex.EncodeToString(predictableNameHash)
			//tempChartArchive := filepath.Join(os.TempDir(), "kod-"+predictableName+".tgz")
			tmpSubFolderName := "kod-package-" + predictableName
			chartFolder = filepath.Join(os.TempDir(), tmpSubFolderName)

			//chartName := ""
			//chartVersion := ""

			// Fetch chart
			if strings.HasPrefix(chartRegistry, "https://") {
				// We have a https helm3 registry URL. Name of the chart is in the -c option
				// Now fetch the chart from a HTTPS URL (Public with no auth)

				chartFolder = filepath.Join(chartFolder, chartPath)
				_, err := os.Stat(chartFolder)
				if err == nil {
					fmt.Println("Helm chart download folder already exists. Skipping.", chartFolder)
				} else {

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
					var helmPullOutput internal.SaveOutput
					helmPullOutput.Prefix = "\xF0\x9F\x8C\x90 "
					helmPullExec.Stdin = os.Stdin
					helmPullExec.Stdout = &helmPullOutput
					helmPullExec.Stderr = os.Stderr

					// Execute the command
					err = helmPullExec.Run()
					if err != nil {
						fmt.Println("Error executing helm pull:", err, "details:", helmPullOutput.String())
						os.Exit(1)
					}

					// Note that helm pull uses the chart name as a subfolder
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

				}

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
			// Source is a tgz file
			if strings.HasSuffix(chartPath, ".tgz") || strings.HasSuffix(chartPath, ".tar.gz") {
				fmt.Println("Using local Helm chart archive tgz file", chartPath)

				predNameString := chartPath
				h := sha256.New()
				h.Write([]byte(predNameString))
				predictableNameHash := h.Sum(nil)
				predictableName := hex.EncodeToString(predictableNameHash)

				tmpSubFolderName := "kod-package-" + predictableName
				tmpFolderPath := filepath.Join(os.TempDir(), tmpSubFolderName)
				fmt.Println("Unpacking to temporary chart folder", tmpFolderPath)
				err := os.MkdirAll(tmpFolderPath, os.ModePerm)
				if err != nil {
					fmt.Println("Error creating temporary chart unarchive folder", err)
					os.Exit(1)
				}

				// Unarchive the folder so we can introspect it
				tmpFolder, err := internal.Unarchive(context.Background(), chartPath, archives.Gz{}, archives.Tar{},
					tmpFolderPath)
				if err != nil {
					fmt.Println("Error unarchiving source chart", err)
					os.Exit(1)
				}

				// Now find the first (and only) subfolder within this folder, as that's the source chart name
				subfolder, err := internal.GetFirstSubfolder(context.Background(), chartFolder)
				if err != nil {
					fmt.Println("Error getting subfolder of unpackaged Chart tgz", err, "Malformed tgz file?")
					os.Exit(1)
				}
				chartFolder = filepath.Join(tmpFolder, subfolder)
				chartPath = subfolder
			}
			// source is a local unpacked chart folder (which may or may not have just been created from a packaged tgz file)
			fmt.Println("Using local unpacked chart folder:", chartPath)
		} // END if remote else local chart reference and copy and unpack

		//var containers []types.ContainerImage
		absChartFolder, err := filepath.Abs(chartFolder)
		if err != nil {
			fmt.Println("Error getting absolute chart folder:", chartFolder, "error:", err)
			os.Exit(1)
		}
		topLevelResult := types.HelmChartProcessingResult{}
		folderName, chartDef, err := internal.ProcessChartFolder("", true, absChartFolder, true, &topLevelResult)

		containerList := types.ContainerImageList{}
		internal.PopulateContainerList(&containerList, &topLevelResult)

		//tempPath := filepath.Join(os.TempDir(), folderName)
		tempPath := folderName // already created for us by above call

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
		for ctrIdx, ctr := range containerList.Containers {
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
				imageUnversionedPath := "docker://" + ctr.Registry + "/" + ctr.Repository
				imagePath := imageUnversionedPath
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
				// if this fails, try version prepended with 'v' (Thank kube-state-metrics for this workaround)
				if err != nil {
					fmt.Println("Error inspecting skopeo image:", ctrFile, "at url:", imagePath, "error:", err, "details:", skopeoInspectOutput.String())
					fmt.Println("Attempting workaround of prepending version with 'v'...")
					skopeoInspectOutput.Clear()
					imagePath = imageUnversionedPath + ":v" + ctr.Tag
					skopeoInspectExec = exec.Command("skopeo", "inspect", imagePath)
					skopeoInspectExec.Stdin = os.Stdin
					skopeoInspectExec.Stdout = &skopeoInspectOutput
					skopeoInspectExec.Stderr = os.Stderr
					err = skopeoInspectExec.Run()
					if err == nil {
						warnings = append(warnings, types.ComplianceWarning{
							Type:      types.ReferencedTypeHelmChart,
							Reference: chartDef.Name + ":" + chartDef.Version,
							Message: fmt.Sprintf("You are incorrectly using a container tag version for '%s' of '%s' when it should be '%s'. Please fix this.",
								ctr.Repository, ctr.Tag, "v"+ctr.Tag),
							Detail: skopeoInspectOutput.String(),
						})
						ctr.Tag = "v" + ctr.Tag
						ctrFile = filepath.Join(cf, ctr.Repository, ctr.Tag+".tar")
					}
				}
				// if this fails, try 'latest' (Thank Bitnami for this workaround)
				if err != nil {
					fmt.Println("Error inspecting skopeo image:", ctrFile, "at url:", imagePath, "error:", err, "details:", skopeoInspectOutput.String())
					fmt.Println("Attempting workaround of using 'latest' version...")
					skopeoInspectOutput.Clear()
					imagePath = imageUnversionedPath + ":latest"
					skopeoInspectExec = exec.Command("skopeo", "inspect", imagePath)
					skopeoInspectExec.Stdin = os.Stdin
					skopeoInspectExec.Stdout = &skopeoInspectOutput
					skopeoInspectExec.Stderr = os.Stderr
					err = skopeoInspectExec.Run()
					if err == nil {
						warnings = append(warnings, types.ComplianceWarning{
							Type:      types.ReferencedTypeHelmChart,
							Reference: chartDef.Name + ":" + chartDef.Version,
							Message: fmt.Sprintf("You are incorrectly using a container tag version for '%s' of '%s' when it should be '%s'. Please fix this",
								ctr.Repository, ctr.Tag, "latest"),
							Detail: skopeoInspectOutput.String(),
						})
						ctr.Tag = "latest"
						ctrFile = filepath.Join(cf, ctr.Repository, ctr.Tag+".tar")
					}
				}
				// if this fails, try 'master' (Thank prometheus-community windows-exporter for this workaround)
				if err != nil {
					fmt.Println("Error inspecting skopeo image:", ctrFile, "at url:", imagePath, "error:", err, "details:", skopeoInspectOutput.String())
					fmt.Println("Attempting workaround of using 'master' version...")
					skopeoInspectOutput.Clear()
					imagePath = imageUnversionedPath + ":master"
					skopeoInspectExec = exec.Command("skopeo", "inspect", imagePath)
					skopeoInspectExec.Stdin = os.Stdin
					skopeoInspectExec.Stdout = &skopeoInspectOutput
					skopeoInspectExec.Stderr = os.Stderr
					err = skopeoInspectExec.Run()
					if err == nil {
						warnings = append(warnings, types.ComplianceWarning{
							Type:      types.ReferencedTypeHelmChart,
							Reference: chartDef.Name + ":" + chartDef.Version,
							Message: fmt.Sprintf("You are incorrectly using a container tag version for '%s' of '%s' when it should be '%s'. Please fix this.",
								ctr.Repository, ctr.Tag, "master"),
							Detail: skopeoInspectOutput.String(),
						})
						ctr.Tag = "master"
						ctrFile = filepath.Join(cf, ctr.Repository, ctr.Tag+".tar")
					}
				}
				// If this fails, give up
				var inspectWorked = true
				if err != nil {
					fmt.Println("Error inspecting skopeo image:", ctrFile, "at url:", imagePath, "error:", err, "details:", skopeoInspectOutput.String())
					warnings = append(warnings, types.ComplianceWarning{
						Type:      types.ReferencedTypeHelmChart,
						Reference: chartDef.Name + ":" + chartDef.Version,
						Message: fmt.Sprintf("You are incorrectly using a container tag version for '%s' of '%s' when it should be '%s'. Please fix this.",
							ctr.Repository, ctr.Tag, "master"),
					})
					fmt.Println("WARNING: We are continuing package creation without this container. Your deployments may fail as a result, depending upon the helm chart's internal logic.")
					//os.Exit(1)
					inspectWorked = false
				}
				if inspectWorked {
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
							warnings = append(warnings, types.ComplianceWarning{
								Type:      types.ReferencedTypeContainerImage,
								Reference: ctr.Repository + ":" + ctr.Tag + "@" + ctr.Digest,
								Message: fmt.Sprintf("You are incorrectly using a container digest for '%s' of '%s' when it should be '%s'. Please fix this.",
									ctr.Repository, ctr.Digest, inspectData.Digest),
							})
							fmt.Println("WARNING: Declared digest for container in helm and in registry are different. Helm digest:", ctr.Digest, "registry digest:", inspectData.Digest, "Continuing anyway. Overwriting Digest value.")
						}
					}
					// Set container digest
					ctr.Digest = inspectData.Digest
					containerList.Containers[ctrIdx] = ctr // sets the value

					// Update incase the change in Tag means ctrFile needs changing
					// Recheck ctrFile incase the corrected version already exists
					ctrFile = filepath.Join(cf, ctr.Repository, ctr.Tag+".tar") // changed so that last filename is the tag version, incase container name and version both have hyphens!
					_, err = os.Stat(ctrFile)
					if err == nil {
						fmt.Println("Container folder exists, skipping skopeo copy to", ctrFile)
					} else {
						// Now invoke skopeo - skopeo copy docker://quay.io/buildah/stable docker-archive:///tmp/kod-redis-1.2.3/containers/docker.io/redis/1.2.3
						srcPath := "docker://" + ctr.Registry + "/" + ctr.Repository + ":" + ctr.Tag
						if ctr.Digest != "" {
							srcPath = "docker://" + ctr.Registry + "/" + ctr.Repository + "@" + ctr.Digest
						}
						skopeoExec := exec.Command("skopeo", "copy", srcPath, "docker-archive:"+ctrFile, "--retry-times", "5")
						var skopeoOutput internal.SaveOutput
						skopeoOutput.Prefix = "\xF0\x9F\x9A\xA2 "
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
								warnings = append(warnings, types.ComplianceWarning{
									Type:      types.ReferencedTypeContainerImage,
									Reference: ctr.Repository + ":" + ctr.Tag + "@" + ctr.Digest,
									Message: fmt.Sprintf("Your container manifest uses a tag for '%s' of '%s', with no fallback version detected (E.g. latest). Please fix this.",
										ctr.Repository, ctr.Tag),
									Detail: skopeoOutput.String(),
								})
								fmt.Println("WARNING: We are continuing package creation without this container. Your deployments may fail as a result, depending upon the helm chart's internal logic.")
								//os.Exit(1)
							} else {
								warnings = append(warnings, types.ComplianceWarning{
									Type:      types.ReferencedTypeContainerImage,
									Reference: ctr.Repository + ":" + ctr.Tag + "@" + ctr.Digest,
									Message: fmt.Sprintf("Your container manifest uses a tag for '%s' of '%s',  but a fallback of '%s'. Please fix this.",
										ctr.Repository, ctr.Tag, "latest"),
								})

								warnings = append(warnings, types.ComplianceWarning{
									Type:      types.ReferencedTypeContainerImage,
									Reference: ctr.Repository + ":" + ctr.Tag + "@" + ctr.Digest,
									Message: fmt.Sprintf("Your container manifest uses the 'latest' tag for '%s' instead of a valid specific version number or digest. Please fix this.",
										ctr.Repository),
								})

								fmt.Println(fmt.Sprintf("WARNING: We are continuing package creation with this container at the '%s' version, but marked with tag '%s'. Your deployments may fail as a result, depending upon the helm chart's internal logic.",
									"latest", ctr.Tag))
							}
						}
						// TODO verify sha, if it exists and we didn't download using the SHA itself
					}
				}
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
			Containers: containerList.Containers,
		}
		pkgBytes, err := yaml.Marshal(pkgDef)
		if err != nil {
			fmt.Println("Error marshalling kod-package.yaml", err)
			os.Exit(1)
		}
		packageFilePath := filepath.Join(tempPath, "kod-package.yaml")
		err = os.WriteFile(packageFilePath, pkgBytes, os.ModePerm)
		if err != nil {
			fmt.Println("Error writing kod-package.yaml", err)
			os.Exit(1)
		}
		fmt.Println("Written package definition to temporary file", packageFilePath)

		fmt.Println("Writing compliance report file...")
		warnBytes, err := yaml.Marshal(types.ComplianceReport{
			GeneratedBy:        "kod",
			Generated:          time.Now().Format(time.RFC3339),
			ComplianceWarnings: warnings,
		})
		if err != nil {
			fmt.Println("Error marshalling kod-compliance-report.yaml", err)
			os.Exit(1)
		}
		warnFilePath := filepath.Join(tempPath, "kod-compliance-report.yaml")
		err = os.WriteFile(warnFilePath, warnBytes, os.ModePerm)
		if err != nil {
			fmt.Println("Error writing kod-compliance-report.yaml", err)
			os.Exit(1)
		}
		fmt.Println("Written compliance report to temporary file", warnFilePath)

		fmt.Println("\xF0\x9F\x93\xA6 Creating final package archive (.kodpkg) ...")

		//firstFolder := filepath.Base(folderName)
		firstFolder := "kod-" + chartDef.Name + "-" + chartDef.Version
		packagePath := filepath.Join(os.TempDir(), "kod-"+chartDef.Name+"-"+chartDef.Version+".kodpkg")

		// Using tar command as a much faster option than the archives module
		// Also added progress dots as per DoUnpack command use of tar
		// Note: The below escaping doesn't work when passing directly to exec, you need to use bash's escaping
		//tempPathForTar := strings.ReplaceAll(tempPath, ".", "\\.")
		//tempPathForTar := strings.ReplaceAll(tempPath, "/", "\\\\/")
		//firstFolderForTar := strings.ReplaceAll(firstFolder, ".", "\\.")
		//firstFolderForTar := strings.ReplaceAll(firstFolder, "/", "\\\\/")
		//tarExec := exec.Command("tar", "cJf", packagePath, "--checkpoint=1000", "--checkpoint-action=dot", "--transform", "'s,^"+tempPathForTar[3:]+","+firstFolderForTar+",'", tempPath)
		tarExec := exec.Command("bash", "-c", "tar cJf "+packagePath+" --checkpoint=1000 --checkpoint-action=dot --transform 's,^"+tempPath[1:]+","+firstFolder+",' -C / "+tempPath[1:])
		// Note: -C / and tempPath[1:] above means you don't get the 'removing /' message in tar messing with the positioning of the ... progress bar with the emoji
		fmt.Println("Executing", tarExec.String())
		var tarOutput internal.SaveOutput
		//tarOutput.Prefix = "\xF0\x9F\x93\x82 " // don't do this as we'll get one per 'dot' in the progress bar!
		fmt.Print("  \xF0\x9F\x93\x82 ")
		tarExec.Stdin = os.Stdin
		tarExec.Stdout = &tarOutput
		tarExec.Stderr = os.Stderr
		err = tarExec.Run()
		//if err != nil {
		//	return fmt.Errorf("unpacking kodpkg archive failed: %s", tarOutput.String())
		//}

		// Now package the temp folder as a tar.xz but with the kodpkg extension
		//files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
		//	tempPath: firstFolder,
		//})
		//if err != nil {
		//	fmt.Println("Error specifying kod package archive", err)
		//	os.Exit(1)
		//}

		// create the output file we'll write to
		//out, err := os.Create(packagePath)
		//if err != nil {
		//	fmt.Println("Error creating kod package archive", err)
		//	os.Exit(1)
		//}
		//defer out.Close()

		// we can use the CompressedArchive type to gzip a tarball
		// (since we're writing, we only set Archival, but if you're
		// going to read, set Extraction)
		//format := archives.CompressedArchive{
		//	Compression: archives.Xz{},
		//	Archival:    archives.Tar{},
		//}
		//
		//// create the archive
		//err = format.Archive(context.Background(), out, files)
		if err != nil {
			fmt.Println("Error writing kod package archive", err, "details:", tarOutput.String())
			os.Exit(1)
		}

		if len(warnings) > 0 {
			fmt.Println("WARNING: Kod package archive was created with compliance warnings. See this file for details:",
				warnFilePath)
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
