/*
Copyright © 2026 Kod project Contributors
*/
package types

/*
 * Holds an individual OCI Layer metadata
 */
type LayerData struct {
	MIMEType    string            `yaml:"MIMEType"`
	Digest      string            `yaml:"Digest"`
	Size        int64             `yaml:"Size"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

/*
 * Top level struct holding the results from the skopeo inspect CONTAINERURL command.
 */
type SkopeoInspectResult struct {
	Name          string            `yaml:"Name"`
	Digest        string            `yaml:"Digest"`
	RepoTags      []string          `yaml:"RepoTags"`
	Created       string            `yaml:"Created"`
	DockerVersion string            `yaml:"DockerVersion"`
	Labels        map[string]string `yaml:"Labels"`
	Architecture  string            `yaml:"Architecture"`
	OS            string            `yaml:"Os"`
	Layers        []string          `yaml:"Layers"`
	LayersData    []LayerData       `yaml:"LayersData"`
	Env           []string          `yaml:"Env"`
}
