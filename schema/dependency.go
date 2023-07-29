package schema

type Dependency struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type Dist struct {
	Tarball string `json:"tarball"`
}

type PackageVersionInfo struct {
	Version      string            `json:"version"`
	Dist         Dist              `json:"dist"`
	Dependencies map[string]string `json:"dependencies"` // Add this field
}
