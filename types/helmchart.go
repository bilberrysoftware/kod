/*
Copyright © 2026 Kod project Contributors
*/
package types

/*
 * Helm chart dependency information as found in Chart.yaml
 */
type HelmChartDependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
}

/*
 * Top level Chart.yaml struct for the items we care about in Kod only
 */
type HelmChart struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Sources      []string              `yaml:"sources"`
	ApiVersion   string                `yaml:"apiVersion"`
	AppVersion   string                `yaml:"appVersion"`
	Dependencies []HelmChartDependency `yaml:"dependencies"`
}
