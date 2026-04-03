/*
Copyright © 2026 Kod project Contributors
*/
package types

type ContainerImage struct {
	Registry   string `yaml:"registry"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	Digest     string `yaml:"digest"`
}
