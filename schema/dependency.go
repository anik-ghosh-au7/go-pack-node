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

type LockDependency struct {
	Version       string            `json:"version"`
	ParentPackage string            `json:"parentPackage"`
	Resolved      string            `json:"resolved"`
	Dependencies  map[string]string `json:"dependencies"`
}

type DependencyLock struct {
	Dependencies map[string]*LockDependency `json:"dependencies"`
}
