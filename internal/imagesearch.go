package internal

import (
	"fmt"
	"kod/types"
	"reflect"
	"regexp"
	"strings"
)

/*
 * Find a container image reference directly below the given values.yaml element
 */
func FindContainerImagesByImageChildValues(chartDef types.HelmChart, parentPath string, parent *map[string]interface{}, containers *[]types.ContainerImage, hints *types.HintsFile) error {
	vImage := *parent

	hint := types.ContainerImageHint{
		ParentPath: parentPath,
		//RegistryPath:   vImageRegistryStr,
		//RepositoryPath  vImageRepositoryStr
		//TagPath:        vImageTagStr,
		//DigestPath:     vImageDigestStr,
		//PackagedImage:  ctrObj,
	}
	// check for 'registry' first
	vImageRegistry := vImage["registry"]     // TODO change this to be more dynamic in future versions
	vImageRepository := vImage["repository"] // TODO change this to be more dynamic in future versions
	vImageRegistryStr := ""
	if vImageRepository != nil {
		vImageRepositoryStr := strings.TrimSpace(vImageRepository.(string))
		if vImageRepositoryStr != "" {
			hint.RepositoryPath = "repository" // TODO change this to be more dynamic in future versions

			vImageDigest := vImage["digest"] // TODO change this to be more dynamic in future versions
			vImageDigestStr := ""
			if vImageDigest == nil {
				// take the last part from repository after the @, if specified, otherwise leave as ""
				idx := strings.LastIndex(vImageRepositoryStr, "@")
				if idx != -1 {
					vImageRepositoryStr = vImageRepositoryStr[:idx]
					vImageDigestStr = vImageRepositoryStr[idx+1:]
					hint.DigestPath = "repository.@"
				}
			} else {
				vImageDigestStr = vImageDigest.(string)
				hint.DigestPath = "digest"
			}
			if vImageRegistry == nil {
				// Get Registry from the first part of repository if it looks like a URL, OR default to docker.io
				idx := strings.Index(vImageRepositoryStr, "/")
				if idx != -1 {
					vImageRegistryStr = vImageRepositoryStr[:idx]
					vImageRepositoryStr = vImageRepositoryStr[idx+1:]
					hint.RegistryPath = "repository./"
				} else {
					// Ensure the first part is NOT an FQDN before we set to docker.io
					strToCompare := vImageRepositoryStr[:idx]
					fqdnRE := regexp.MustCompile("^[a-zA-Z0-9._-]+[a-zA-Z0-9]\\.[a-zA-Z0-9._-]+[a-zA-Z0-9]$")
					if fqdnRE.MatchString(strToCompare) {
						// Still a repository with a registry path in it
						vImageRegistryStr = vImageRepositoryStr[:idx]
						vImageRepositoryStr = vImageRepositoryStr[idx+1:]
						hint.RegistryPath = "repository./"
					} else {
						// vImageRegistry isn't in repository either - so default to docker.io, and assume the override is in repository
						fmt.Println(" - WARNING: registry path is not specified, so setting to docker.io and assuming it's overridden in the repository tag path of:", vImageRepositoryStr)
						vImageRegistryStr = "docker.io"
						hint.RegistryPath = "repository./"
					}
				}
			} else {
				vImageRegistryStr = strings.TrimSpace(vImageRegistry.(string))
				hint.RegistryPath = "registry"
			}
			vImageTag := vImage["tag"] // TODO change this to be more dynamic in future versions
			vImageTagStr := ""
			if vImageTag == nil {
				// Get Tag from the last part of the repository (now that we've removed digest
				idx := strings.LastIndex(vImageRepositoryStr, ":")
				if idx != -1 {
					vImageTagStr = vImageRepositoryStr[:idx]
					vImageRepositoryStr = vImageRepositoryStr[idx+1:]
					hint.TagPath = "repository.:"
				}
			} else {
				vImageTagStr = strings.TrimSpace(vImageTag.(string))
				hint.TagPath = "tag"
			}
			// Last catch all for tag
			if vImageTagStr == "" {
				fmt.Println("WARNING: no version tag found for container. Defaulting to 'Chart.appVersion' for", vImageRepositoryStr)
				//vImageTagStr = "latest"
				// default to Chart.AppVersion when this is blank in the values file
				vImageTagStr = chartDef.AppVersion
				// TODO Consider specifying a target version of "sha256-SHAVALUE" when this happens, to avoid CIS Benchmark issues on deployment
			}
			fmt.Println(fmt.Sprintf("- Found container image: '%s/%s:%s@%s' at path '%s'", vImageRegistryStr, vImageRepositoryStr, vImageTagStr, vImageDigestStr, parentPath))
			ctrObj := types.ContainerImage{
				Registry:   vImageRegistryStr,
				Repository: vImageRepositoryStr,
				Tag:        vImageTagStr,
				Digest:     vImageDigestStr,
			}
			*containers = append(*containers, ctrObj) // TODO verify that this works OK with dereferencing
			// Save the value mappings of this information so we can override the correct parameters on deployment of the package

			hint.PackagedImage = ctrObj
			hints.Images = append(hints.Images, hint)
		}
	}

	return nil
}

/*
 * Search for a potential container image parent by matching its name to .*[iI]mage$, and trying to add containers from within it.
 * To begin a search at the top element, pass parentPath as the empty string ""
 */
func FindContainerImagesByImageTagSearch(chartDef types.HelmChart, parentPath string, valuesMap *map[string]interface{}, containers *[]types.ContainerImage, hints *types.HintsFile) error {
	var imageTag = regexp.MustCompile(`.*[iI]mage$`)
	targetType := reflect.TypeOf(*valuesMap)
	for key, value := range *valuesMap {
		if value != nil {
			if reflect.TypeOf(value) == targetType {
				// Calculate value's full key path
				valuePath := parentPath
				if valuePath != "" {
					valuePath += "."
				}
				valuePath += key

				// Check if this key matches, and don't process sub keys if it does
				if imageTag.MatchString(key) {
					// This is possible if the element exists but has a nil value
					el := value.(map[string]interface{})
					err := FindContainerImagesByImageChildValues(chartDef, valuePath, &el, containers, hints)
					if err != nil {
						return err
					}

					// Check for imagePullSecrets under here too
					ipsh := types.SecretHint{}
					ipsh.PackagedSecrets = []types.SecretReference{}
					ipsEl := el["imagePullSecrets"]
					if ipsEl != nil {
						// Option 1. It's an array of strings as per K8s specification
						// This is if it has a value specified. It will be an empty interface{} if blank (which is the norm)
						if reflect.TypeOf(ipsEl) == reflect.TypeOf([]string{}) {
							ips := ipsEl.([]string)
							for _, secretRef := range ips {
								ipsh.PackagedSecrets = append(ipsh.PackagedSecrets, types.SecretReference{Name: secretRef})
							}
							//} else {
							//	hints.ImagePullSecrets.PackagedSecrets = []string{}
							ipsh.SecretArrayPath = valuePath + ".imagePullSecrets"
							hints.ImagePullSecrets = append(hints.ImagePullSecrets, ipsh)
						}
						// Option 2. It's a single secret name with an enabling flag
						// Ensure type is map[string]interface{}{}, otherwise check for 'enabled' boolean and 'name'(which may not exist) - NiFiKop operator
						if reflect.TypeOf(ipsEl) == reflect.TypeOf(map[string]interface{}{}) {
							ipsNonArray := ipsEl.(map[string]interface{})

							nameEl := ipsNonArray["name"]
							if nameEl != nil {
								ipsh.PackagedSecrets = append(ipsh.PackagedSecrets, types.SecretReference{Name: nameEl.(string)})
							}
							ipsh.SecretNamePath = valuePath + ".imagePullSecrets.name"

							enabledEl := ipsNonArray["enabled"]
							if enabledEl != nil {
								ipsh.EnabledFlagPath = valuePath + ".imagePullSecrets.enabled"
							}

							hints.ImagePullSecrets = append(hints.ImagePullSecrets, ipsh)
						}
					}
				} else {
					// Otherwise, process sub keys (depth first search
					// Qn: Is there any circumstance where an image parent may be within another tag whose parent matches the regexp too?
					valueMap := value.(map[string]interface{})
					err := FindContainerImagesByImageTagSearch(chartDef, valuePath, &valueMap, containers, hints)
					if err != nil {
						return err
					}
				}

			}
		}
	}

	return nil
}

func FindContainerImagesByGlobalRegistryAndComponentName(chartDef types.HelmChart, parentPath string, valuesMap *map[string]interface{}, containers *[]types.ContainerImage, hints *types.HintsFile) error {
	// Find global components first
	vImage := *valuesMap
	vImageRegistry := vImage["imageRegistry"]     // TODO change this to be more dynamic in future versions
	vImageRepository := vImage["imageRepository"] // TODO change this to be more dynamic in future versions
	vImageRegistryStr := ""
	vImageRepositoryStr := ""
	repositoryField := "imageNamespace"
	if vImageRegistry != nil {
		vImageRegistryStr = strings.TrimSpace(vImageRegistry.(string))
	}
	if vImageRepository != nil {
		vImageRepositoryStr = strings.TrimSpace(vImageRepository.(string))
		repositoryField = "imageRepository"
	} else {
		// try image namespace instead
		vImageRepository = vImage["imageNamespace"]
		if nil != vImageRepository {
			vImageRepositoryStr = strings.TrimSpace(vImageRepository.(string))
		}
	}
	if vImageRegistryStr == "" || vImageRepositoryStr == "" {
		// Not a valid match, return
		fmt.Println(" - Evaluation looking for top level imageRegistry AND (imageRepository OR imageNamespace) is negative. Skipping container search for this discovery method.")
		return nil
	}
	// Then search for COMPONENT.image.name
	// TODO look for non blank overrides, COMPONENT.image.tag and COMPONENT.image.digest, or default to Chart.AppVersion
	return FindContainerImageByComponentNameSearchGivenPath(chartDef, parentPath, valuesMap, "imageRegistry", vImageRegistryStr, repositoryField, vImageRepositoryStr, containers, hints)
}

/*
 * This looks for a top level imageRegistry AND (imageRepository OR imageNamespace), and then recursively searches
 * child elements for the specific image name, tag, and digest.
 * Used by cert-manager at least
 */
func FindContainerImageByComponentNameSearchGivenPath(chartDef types.HelmChart, parentPath string, valuesMap *map[string]interface{}, registryFieldPath string, vImageRegistryStr string, repositoryFieldPath string, vImageRepositoryStr string, containers *[]types.ContainerImage, hints *types.HintsFile) error {
	targetType := reflect.TypeOf(*valuesMap)
	for key, value := range *valuesMap {
		if value != nil {
			if reflect.TypeOf(value) == targetType {
				// Calculate value's full key path
				valuePath := parentPath
				if valuePath != "" {
					valuePath += "."
				}
				valuePath += key

				// Check if this key matches, and don't process sub keys if it does
				if key == "image" {
					// This is possible if the element exists but has a nil value
					el := value.(map[string]interface{})
					nameElement := el["name"]
					digestElement := el["digest"]
					tagElement := el["tag"]
					nameStr := ""
					vImageDigestStr := ""
					vImageTagStr := ""
					// Look for 'name'
					if nameElement == nil {
						fmt.Println(" - Name element is missing, skipping element with content:", el)
						continue
					}
					hint := types.ContainerImageHint{
						ParentPath:     "",
						RegistryPath:   registryFieldPath,
						RepositoryPath: repositoryFieldPath + "/" + valuePath + ".name",
						//TagPath:        vImageTagStr,
						//DigestPath:     vImageDigestStr,
						//PackagedImage:  ctrObj,
					}
					nameStr = nameElement.(string)

					// Look for 'digest' and 'tag' also
					if digestElement != nil {
						vImageDigestStr = digestElement.(string)
						hint.DigestPath = valuePath + ".digest"
					}
					if tagElement != nil {
						vImageTagStr = tagElement.(string)
						hint.TagPath = valuePath + ".tag"
					}
					if vImageTagStr == "" {
						fmt.Println("WARNING: no version tag found for container. Defaulting to 'Chart.appVersion' for", vImageRepositoryStr)
						// default to Chart.AppVersion when this is blank in the values file
						vImageTagStr = chartDef.AppVersion
						// TODO ensure this doesn't break an installation if it doesn't exist
						hint.TagPath = valuePath + ".tag"
					}

					fmt.Println(fmt.Sprintf("- Found container image: '%s/%s/%s:%s@%s' at path '%s'", vImageRegistryStr, vImageRepositoryStr, nameStr, vImageTagStr, vImageDigestStr, parentPath))

					ctrObj := types.ContainerImage{
						Registry:   vImageRegistryStr,
						Repository: vImageRepositoryStr + "/" + nameStr,
						Tag:        vImageTagStr,
						Digest:     vImageDigestStr,
					}
					*containers = append(*containers, ctrObj)
					// Save the value mappings of this information so we can override the correct parameters on deployment of the package

					hint.PackagedImage = ctrObj
					hints.Images = append(hints.Images, hint)

				} else {
					// Otherwise, process sub keys (depth first search
					// Qn: Is there any circumstance where an image parent may be within another tag whose parent matches the regexp too?
					valueMap := value.(map[string]interface{})
					err := FindContainerImageByComponentNameSearchGivenPath(chartDef, valuePath, &valueMap, registryFieldPath, vImageRegistryStr, repositoryFieldPath, vImageRepositoryStr, containers, hints)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
