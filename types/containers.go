/*
Copyright © 2026 Kod project Contributors
*/
package types

/*
 * Specifies the definition of a container image as processed by kod during packaging.
 * Note that tag and digest are required at this stage, because a container must have a human readable version
 * (even if only 'latest') and the (usually 'sha256:' prefixed) hash must also be known upon packaging.
 * Registry can only be known if it can be extracted from repository if that is the only image name element provided.
 * We could assumed that the first domain name only is the registry, but some may exist under a sub path and not at
 * the root '/' path on the server, thus we do not assume this common, but not universal, fact.
 */
type ContainerImage struct {
	Registry   string `yaml:"registry,omitempty"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	Digest     string `yaml:"digest"`
}
