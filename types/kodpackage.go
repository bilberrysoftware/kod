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

/**
 * Acts as the top level of the CHARTNAME-CHARTVER-hints.yaml file
 */
type HintsFile struct {
	ChartRef HelmChartReference   `yaml:"chartRef"`
	Hints    []ContainerImageHint `yaml:"hints,omitempty"`
}
