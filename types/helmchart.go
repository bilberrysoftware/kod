/*
Copyright © 2026 Kod project Contributors
*/
package types

type HelmChartDependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
}
type HelmChart struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Sources      []string              `yaml:"sources"`
	ApiVersion   string                `yaml:"apiVersion"`
	AppVersion   string                `yaml:"appVersion"`
	Dependencies []HelmChartDependency `yaml:"dependencies"`
}
