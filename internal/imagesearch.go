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
