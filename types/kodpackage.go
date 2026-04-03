package types

type Package struct {
	Type       string           `yaml:"type"`
	Name       string           `yaml:"name"`
	Version    string           `yaml:"version"`
	Sources    []string         `yaml:"sources"`
	Containers []ContainerImage `yaml:"containers"`
}
