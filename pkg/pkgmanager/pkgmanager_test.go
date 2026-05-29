package pkgmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNFManagerInterface(t *testing.T) {
	m := New("dnf")
	assert.NotNil(t, m)
	assert.Equal(t, "dnf", m.Name())
}

func TestAPTManagerInterface(t *testing.T) {
	m := New("apt")
	assert.NotNil(t, m)
	assert.Equal(t, "apt", m.Name())
}

func TestPacmanManagerInterface(t *testing.T) {
	m := New("pacman")
	assert.NotNil(t, m)
	assert.Equal(t, "pacman", m.Name())
}

func TestNewUnknownDistro(t *testing.T) {
	m := New("alpine")
	assert.Nil(t, m)
}

func TestDNFInstallArgs(t *testing.T) {
	m := New("dnf")
	args := m.Install([]string{"fish", "vim"})
	assert.Equal(t, []string{"dnf", "install", "-y", "fish", "vim"}, args)
}

func TestDNFUpdateDBArgs(t *testing.T) {
	m := New("dnf")
	args := m.UpdateDB()
	assert.Equal(t, []string{"dnf", "makecache"}, args)
}

func TestAPTInstallArgs(t *testing.T) {
	m := New("apt")
	args := m.Install([]string{"fish", "vim"})
	assert.Equal(t, []string{"apt-get", "install", "-y", "fish", "vim"}, args)
}

func TestAPTUpdateDBArgs(t *testing.T) {
	m := New("apt")
	args := m.UpdateDB()
	assert.Equal(t, []string{"apt-get", "update"}, args)
}

func TestPacmanInstallArgs(t *testing.T) {
	m := New("pacman")
	args := m.Install([]string{"fish", "vim"})
	assert.Equal(t, []string{"pacman", "-S", "--noconfirm", "fish", "vim"}, args)
}

func TestPacmanUpdateDBArgs(t *testing.T) {
	m := New("pacman")
	args := m.UpdateDB()
	assert.Equal(t, []string{"pacman", "-Sy"}, args)
}

func TestResolveFromDistroFedora(t *testing.T) {
	m := ResolveFromDistro("fedora")
	assert.NotNil(t, m)
	assert.Equal(t, "dnf", m.Name())
}

func TestResolveFromDistroRHEL(t *testing.T) {
	m := ResolveFromDistro("rhel")
	assert.NotNil(t, m)
	assert.Equal(t, "dnf", m.Name())
}

func TestResolveFromDistroUbuntu(t *testing.T) {
	m := ResolveFromDistro("ubuntu")
	assert.NotNil(t, m)
	assert.Equal(t, "apt", m.Name())
}

func TestResolveFromDistroDebian(t *testing.T) {
	m := ResolveFromDistro("debian")
	assert.NotNil(t, m)
	assert.Equal(t, "apt", m.Name())
}

func TestResolveFromDistroArch(t *testing.T) {
	m := ResolveFromDistro("arch")
	assert.NotNil(t, m)
	assert.Equal(t, "pacman", m.Name())
}

func TestResolveFromDistroUnknown(t *testing.T) {
	m := ResolveFromDistro("suse")
	assert.Nil(t, m)
}

func TestDNFListInstalled(t *testing.T) {
	m := New("dnf")
	args := m.ListInstalled()
	assert.Equal(t, []string{"dnf", "list", "installed"}, args)
}

func TestAPTListInstalled(t *testing.T) {
	m := New("apt")
	args := m.ListInstalled()
	assert.Equal(t, []string{"dpkg", "-l"}, args)
}

func TestPacmanListInstalled(t *testing.T) {
	m := New("pacman")
	args := m.ListInstalled()
	assert.Equal(t, []string{"pacman", "-Q"}, args)
}

func TestDNFQuery(t *testing.T) {
	m := New("dnf")
	args := m.Query("fish")
	assert.Equal(t, []string{"rpm", "-q", "fish"}, args)
}

func TestAPTQuery(t *testing.T) {
	m := New("apt")
	args := m.Query("fish")
	assert.Equal(t, []string{"dpkg", "-s", "fish"}, args)
}

func TestPacmanQuery(t *testing.T) {
	m := New("pacman")
	args := m.Query("fish")
	assert.Equal(t, []string{"pacman", "-Qi", "fish"}, args)
}

func TestDNFRawInstallCmdNoExtraAllocs(t *testing.T) {
	m := New("dnf")
	args := m.Install(nil)
	assert.Equal(t, []string{"dnf", "install", "-y"}, args)
}

func TestAPTInstallEmpty(t *testing.T) {
	m := New("apt")
	args := m.Install([]string{})
	assert.Equal(t, []string{"apt-get", "install", "-y"}, args)
}

func TestPacmanInstallSinglePkg(t *testing.T) {
	m := New("pacman")
	args := m.Install([]string{"fish"})
	assert.Equal(t, []string{"pacman", "-S", "--noconfirm", "fish"}, args)
}
