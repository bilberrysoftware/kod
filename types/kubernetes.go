package types

/*
 * Represents the output of kubectl get ns -oyaml MYNAMESPACE
 * See kubectl.go and WaitForNamespaceToExist
 */
type K8sNamespace struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name              string `yaml:"name"`
		CreationTimestamp string `yaml:"creationTimestamp"`
	} `yaml:"metadata"`
	Status struct {
		Phase string `yaml:"phase"`
	} `yaml:"status"`
}
