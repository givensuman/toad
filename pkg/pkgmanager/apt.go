package pkgmanager

type APT struct{}

func (m *APT) Name() string { return "apt" }

func (m *APT) Install(pkgs []string) []string {
	args := []string{"apt-get", "install", "-y"}
	return append(args, pkgs...)
}

func (m *APT) UpdateDB() []string {
	return []string{"apt-get", "update"}
}

func (m *APT) ListInstalled() []string {
	return []string{"dpkg", "-l"}
}

func (m *APT) Query(pkg string) []string {
	return []string{"dpkg", "-s", pkg}
}
