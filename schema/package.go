package schema

type Dist struct {
	Tarball string `json:"tarball"`
}

type DistTags struct {
	Latest string `json:"latest"`
}

type VersionInfo struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Main         string            `json:"main"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	Dist         Dist              `json:"dist"`
}

type PackageInfo struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	DistTags    DistTags                `json:"dist-tags"`
	Versions    map[string]*VersionInfo `json:"versions"`
	Time        map[string]string       `json:"time"`
}

type PackageVersionInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Main         string            `json:"main"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	Dist         Dist              `json:"dist"`
}
