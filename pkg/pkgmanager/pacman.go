package pkgmanager

type PacMan struct{}

func (m *PacMan) Name() string { return "pacman" }

func (m *PacMan) Install(pkgs []string) []string {
	args := []string{"pacman", "-S", "--noconfirm"}
	return append(args, pkgs...)
}

func (m *PacMan) UpdateDB() []string {
	return []string{"pacman", "-Sy"}
}

func (m *PacMan) ListInstalled() []string {
	return []string{"pacman", "-Q"}
}

func (m *PacMan) Query(pkg string) []string {
	return []string{"pacman", "-Qi", pkg}
}
