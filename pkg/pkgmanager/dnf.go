package pkgmanager

type DNF struct{}

func (m *DNF) Name() string { return "dnf" }

func (m *DNF) Install(pkgs []string) []string {
	args := []string{"dnf", "install", "-y"}
	return append(args, pkgs...)
}

func (m *DNF) UpdateDB() []string {
	return []string{"dnf", "makecache"}
}

func (m *DNF) ListInstalled() []string {
	return []string{"dnf", "list", "installed"}
}

func (m *DNF) Query(pkg string) []string {
	return []string{"rpm", "-q", pkg}
}
