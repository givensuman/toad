package declaration

type Mount struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"readonly"`
}

type InitHooks struct {
	PostCreate []string `yaml:"post-create"`
	PostStart  []string `yaml:"post-start"`
}

type Declaration struct {
	Image     string              `yaml:"image,omitempty"`
	Distro    string              `yaml:"distro,omitempty"`
	Release   string              `yaml:"release,omitempty"`
	Container string              `yaml:"container,omitempty"`
	WithPkgs  map[string][]string `yaml:"with-pkgs,omitempty"`
	WithFlags []string            `yaml:"with-flags,omitempty"`
	Mounts    []Mount             `yaml:"mounts,omitempty"`
	Env       map[string]string   `yaml:"env,omitempty"`
	InitHooks *InitHooks          `yaml:"init-hooks,omitempty"`
}
