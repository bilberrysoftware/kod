package types

const (
	ReferencedTypeHelmChart      string = "HelmChart"
	ReferencedTypeContainerImage string = "ContainerImage"
)

type ComplianceWarning struct {
	Type      string `yaml:"type"`
	Reference string `yaml:"reference"`
	Message   string `yaml:"message"`
	Detail    string `yaml:"detail"`
}

/*
 * Compliance report sent out with every kod package, so consumers know what the likely issues are
 */
type ComplianceReport struct {
	Generated          string              `yaml:"generated"`
	GeneratedBy        string              `yaml:"generatedBy"`
	ComplianceWarnings []ComplianceWarning `yaml:"complianceWarnings"`
}
