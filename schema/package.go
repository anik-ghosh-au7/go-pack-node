package schema

type Dist struct {
	Tarball string `json:"tarball"`
}

type PackageVersionInfo struct {
	Version      string            `json:"version"`
	Dist         Dist              `json:"dist"`
	Dependencies map[string]string `json:"dependencies"`
}
