/*
Copyright © 2026 Kod project Contributors
*/
package types

/*
 * Describes a final packaged kod Package. Just what is included and not how we derived this content.
 */
type Package struct {
	Type       string           `yaml:"type"`
	Name       string           `yaml:"name"`
	Version    string           `yaml:"version"`
	Sources    []string         `yaml:"sources,omitempty"`
	Containers []ContainerImage `yaml:"containers,omitempty"`
}

/*
 * Describes where in the original Helm chart kod can find elements pointing to container images.
 * All must have ParentPath. Empty string is a valid ParentPath and means 'under the root of the values file'.
 * Valid combinations are:-
 * - RepositoryPath only
 * - RepositoryPath and TagPath
 * - RepositoryPath and DigestPath
 * - RepositoryPath and TagPath and DigestPath
 * - RegistryPath and RepositoryPath (shortened)
 * - RegistryPath and RepositoryPath (shortened) and TagPath
 * - RegistryPath and RepositoryPath (shortened) and DigestPath
 * - RegistryPath and RepositoryPath (shortened) and TagPath and DigestPath
 */
type ContainerImageHint struct {
	ParentPath     string         `yaml:"parentPath"`
	RegistryPath   string         `yaml:"registryPath,omitempty"`
	RepositoryPath string         `yaml:"repositoryPath"`
	TagPath        string         `yaml:"tagPath,omitempty"`
	DigestPath     string         `yaml:"digestPath,omitempty"`
	PackagedImage  ContainerImage `yaml:"packagedImage"`
}

/*
 * Basic reference, as used by HintsFile
 */
type HelmChartReference struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type SecretReference struct {
	Name string `yaml:"name"`
}

/*
 * Represents a pointer to an image pull secrets path - which is either an array or a single secret name
 */
type SecretHint struct {
	SecretArrayPath string            `yaml:"secretArrayPath,omitempty"` // if it's an imagePullSecrets array, as per spec
	SecretNamePath  string            `yaml:"secretNamePath,omitempty"`  // if it's a simple string element (just use first el of array below) (E.g. NiFiKop Operator image.imagePullSecrets.name)
	EnabledFlagPath string            `yaml:"enabledFlagPath,omitempty"` // boolean enabling flag for the above. Usually under the same parent as the above
	PackagedSecrets []SecretReference `yaml:"packagedSecrets"`
}

/*
 * Acts as the top level of the CHARTNAME-CHARTVER-hints.yaml file
 */
type HintsFile struct {
	ChartRef         HelmChartReference   `yaml:"chartRef"`
	ImagePullSecrets []SecretHint         `yaml:"imagePullSecrets,omitempty"`
	Images           []ContainerImageHint `yaml:"images,omitempty"`
}

/*
 * Used to represent the results from recursive HelmChart packaging
 */
type HelmChartProcessingResult struct {
	HintsFile  HintsFile                   `yaml:"hintsFile"`
	Containers []ContainerImage            `yaml:"containers"`
	Children   []HelmChartProcessingResult `yaml:"children"`
}

type ContainerImageList struct {
	Containers []ContainerImage `yaml:"containers"`
}
