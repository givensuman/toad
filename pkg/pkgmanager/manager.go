package pkgmanager

type Manager interface {
	Name() string
	Install(pkgs []string) []string
	UpdateDB() []string
	ListInstalled() []string
	Query(pkg string) []string
}

var distroMap = map[string]string{
	"fedora": "dnf",
	"rhel":   "dnf",
	"ubuntu": "apt",
	"debian": "apt",
	"arch":   "pacman",
}

func New(manager string) Manager {
	switch manager {
	case "dnf":
		return &DNF{}
	case "apt":
		return &APT{}
	case "pacman":
		return &PacMan{}
	default:
		return nil
	}
}

func ResolveFromDistro(distro string) Manager {
	manager, ok := distroMap[distro]
	if !ok {
		return nil
	}
	return New(manager)
}
