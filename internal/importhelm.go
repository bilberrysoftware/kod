package internal

import (
	"fmt"
	"kod/types"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

/**
 * Processes a single source chart folder into a temporary location as the main chart folder
 */
func ProcessChartFolder(rootPackageFolder string, isRootPackage bool, chartFolder string, copyRequired bool, resultToPopulate *types.HelmChartProcessingResult) (string, types.HelmChart, error) {
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
	// Default to acting like a subchart copy
	tempPath := rootPackageFolder
	chartAndVersion := fmt.Sprintf("%s-%s", chartDef.Name, chartDef.Version)
	if isRootPackage {
		// Now print what we've found
		fmt.Println(fmt.Sprintf("Packaging chart. name: '%s', version: '%s', appVersion: `%s`",
			chartDef.Name, chartDef.Version, chartDef.AppVersion))
		// Create temporary folder based on chart name and version
		folderName := "kod-package-" + chartAndVersion
		tempPath = filepath.Join(os.TempDir(), folderName)
		rootPackageFolder = tempPath
		err = os.MkdirAll(tempPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating temp folder", err)
			os.Exit(1)
		}
	}

	// Determine the container images required
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
		err = FindContainerImagesByImageChildValues(chartDef, "image", &vImage, &resultToPopulate.Containers, &hints)
		if err != nil {
			fmt.Println("Error finding image child elements in values.yaml", err)
			os.Exit(1)
		}
	}
	// End Option 1.

	// Option 2. Underneath any element in values.yaml with a parent called .*[iI]mage:
	err = FindContainerImagesByImageTagSearch(chartDef, "", &valuesMap, &resultToPopulate.Containers, &hints)
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

		ipsh := types.SecretHint{}
		ipsh.PackagedSecrets = []types.SecretReference{}

		global := globalEl.(map[string]interface{})
		ipsEl := global["imagePullSecrets"]
		if ipsEl != nil {
			// This is if it has a value specified. It will be an empty interface{} if blank (which is the norm)
			if reflect.TypeOf(ipsEl) == reflect.TypeOf([]string{}) {
				ips := ipsEl.([]string)
				for _, secretRef := range ips {
					ipsh.PackagedSecrets = append(ipsh.PackagedSecrets, types.SecretReference{Name: secretRef})
				}
				//} else {
				//	hints.ImagePullSecrets.PackagedSecrets = []string{}
			}
			ipsh.SecretArrayPath = "global.imagePullSecrets"
		}
		hints.ImagePullSecrets = append(hints.ImagePullSecrets, ipsh)
	} else {
		ipsh := types.SecretHint{}
		ipsh.PackagedSecrets = []types.SecretReference{}

		ipsEl := valuesMap["imagePullSecrets"]
		if ipsEl != nil {
			if reflect.TypeOf(ipsEl) == reflect.TypeOf([]string{}) {
				ips := ipsEl.([]string)
				for _, secretRef := range ips {
					ipsh.PackagedSecrets = append(ipsh.PackagedSecrets, types.SecretReference{Name: secretRef})
				}

				//} else {
				//	nameMap := map[string]interface{}{}
				//	hints.ImagePullSecrets.PackagedSecrets = []string{}
			}
			ipsh.SecretArrayPath = "imagePullSecrets"
		}
		hints.ImagePullSecrets = append(hints.ImagePullSecrets, ipsh)
	}

	// Copy the Chart folder into a subfolder
	fmt.Println("Copying the helm chart...")
	chartCopyPath := filepath.Join(rootPackageFolder, "charts", chartAndVersion)
	// Don't copy if folder already exists
	_, err = os.Stat(chartCopyPath)
	if err == nil {
		fmt.Println("Chart exists, skipping copy of", chartAndVersion)
		copyRequired = false
	}
	//} else {
	err = os.MkdirAll(chartCopyPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error copying chart to temporary folder:", chartCopyPath, "error:", err)
		os.Exit(1)
	}
	// Don't copy if we're a subchart, as that's already been done for us
	if copyRequired {
		err = os.CopyFS(chartCopyPath, os.DirFS(chartFolder))
		if err != nil {
			fmt.Println("Error copying chart to temporary folder from:", chartFolder, "to:", chartCopyPath, "error:", err)
			os.Exit(1)
		}
	}
	// For each dependency in MYCHART/charts/MYDEPENDENCY, read its Chart.yaml, and move it to the MYKOD/charts folder, replacing with a sym link to the right version
	// Note: VERSION comes from the PARENT Chart.yaml file only, not the subchart Chart.yaml file
	// Do this recursively for each chart
	// Now pull in each of their container images
	// Note: Always processing sub chart folder even if it already exists, as we need to know the hintsfile for the children
	for _, dep := range chartDef.Dependencies {
		fmt.Println("  Processing dependency:", dep.Name, "version:", dep.Version)
		// create new versioned folder
		newDepNameAndVersion := fmt.Sprintf("%s-%s", dep.Name, dep.Version)
		newDepFolder := filepath.Join(tempPath, "charts", newDepNameAndVersion)
		// TODO verify this folder exists - we know helm pull creates it, but against a flat file system it probably won't be created! We may have to cd MYCHARTFOLDER; helm dependency update to create them
		//err = os.MkdirAll(newDepFolder, os.ModePerm)
		//if err != nil {
		//	fmt.Println("Error creating subchart target folder:", newDepFolder, "error:", err)
		//	os.Exit(1)
		//}

		// Copy into the folder
		depSrcPath := filepath.Join(chartCopyPath, "charts", dep.Name)
		err = os.CopyFS(newDepFolder, os.DirFS(depSrcPath))
		if err != nil {
			fmt.Println("Error copying subchart to temporary folder from:", depSrcPath, "to:", newDepFolder, "error:", err)
			os.Exit(1)
		}
		// Remove the old folder
		err = os.RemoveAll(depSrcPath)
		if err != nil {
			fmt.Println("Error removing subchart from:", depSrcPath, err)
			os.Exit(1)
		}
		// Create a relative symlink
		symLinkRelPath := "../../" + newDepNameAndVersion
		err = os.Symlink(symLinkRelPath, depSrcPath)
		if err != nil {
			fmt.Println("Error creating subchart symlink to temporary folder. Actual path", newDepFolder, "symlink relative path:", symLinkRelPath, "dependency source path:", depSrcPath, "error:", err)
			os.Exit(1)
		}

		// Now process each child folder for its dependent container images (recursive call)
		absPath, err := filepath.Abs(rootPackageFolder)
		if err != nil {
			fmt.Println("Error getting absolute path for chart at:", rootPackageFolder, "error:", err)
			os.Exit(1)
		}
		absSubPath, err := filepath.Abs(newDepFolder)
		if err != nil {
			fmt.Println("Error getting absolute path for subchart at:", newDepFolder, "error:", err)
			os.Exit(1)
		}
		childResult := types.HelmChartProcessingResult{}
		_, _, err = ProcessChartFolder(absPath, false, absSubPath, false, &childResult)
		if err != nil {
			fmt.Println("Error processing helm subchart:", newDepNameAndVersion, "error:", err)
			os.Exit(1)
		}
		resultToPopulate.Children = append(resultToPopulate.Children, childResult)
	}
	//}

	// Copy hints file over
	if isRootPackage {
		// include subcharts in hints file
		FlattenHintsToParent(&hints, resultToPopulate, "")
	}
	hintsFileName := filepath.Join(rootPackageFolder, "charts", chartAndVersion+"-hints.yaml")
	_, err = os.Stat(hintsFileName)
	if err == nil {
		// File exists
		fmt.Println(" - Hints file already exists for chart, skipping. Chart:", chartAndVersion)
	} else {
		fmt.Println(" - Writing hints file for chart:", chartAndVersion)
		fmt.Println(" - Number of container images in hints file:", len(hints.Images))
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

	resultToPopulate.HintsFile = hints

	return rootPackageFolder, chartDef, nil

}

/*
 * Flattens all helm chart results in all children into the root HintsFile specified.
 * This allows deploy to override deep chart dependencies from a single root values file.
 * Note: We only do this for each root chart, we don't generate them for intermediate charts that both are and have subcharts.
 *       This is due to the way helm applies dependency values into subcharts from a single deploy on the root chart.
 */
func FlattenHintsToParent(rootHints *types.HintsFile, result *types.HelmChartProcessingResult, parentPath string) {
	for _, child := range result.Children {
		newParentPath := parentPath + "." + child.HintsFile.ChartRef.Name
		if "" == parentPath {
			newParentPath = child.HintsFile.ChartRef.Name
		}
		FlattenHintsToParent(rootHints, &child, newParentPath)
	}
	if "" != parentPath {
		for _, hint := range result.HintsFile.Images {
			rootHints.Images = append(rootHints.Images, types.ContainerImageHint{
				ParentPath:     parentPath + "." + hint.ParentPath,
				RegistryPath:   hint.RegistryPath,
				RepositoryPath: hint.RepositoryPath,
				TagPath:        hint.TagPath,
				DigestPath:     hint.DigestPath,
				PackagedImage:  hint.PackagedImage,
			})
		}
		for _, secretHint := range rootHints.ImagePullSecrets {
			rootHints.ImagePullSecrets = append(rootHints.ImagePullSecrets, types.SecretHint{
				SecretArrayPath: parentPath + "." + secretHint.SecretArrayPath,
				PackagedSecrets: secretHint.PackagedSecrets,
			})
		}
	}
}

func PopulateContainerList(containerList *types.ContainerImageList, helmResultTree *types.HelmChartProcessingResult) {
	// Add our containers first
	for _, ctr := range helmResultTree.Containers {
		found := false
		for _, img := range containerList.Containers {
			found = found || (ctr.Registry == img.Registry) && (ctr.Repository == img.Repository) &&
				((ctr.Tag != "" && ctr.Tag == img.Tag) || ( /* Implied: ctr.Tag == "" && */ ctr.Digest != "" && ctr.Digest == img.Digest))
		}
		if !found {
			containerList.Containers = append(containerList.Containers, ctr)
		}
	}
	// Now add sub chart images
	for _, child := range helmResultTree.Children {
		PopulateContainerList(containerList, &child)
	}
}
