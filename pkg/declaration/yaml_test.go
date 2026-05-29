package declaration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMinimalDeclaration(t *testing.T) {
	yamlContent := `
distro: fedora
release: "42"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "fedora", decl.Distro)
	assert.Equal(t, "42", decl.Release)
	assert.Empty(t, decl.Image)
}

func TestParseImageDeclaration(t *testing.T) {
	yamlContent := `
image: registry.fedoraproject.org/fedora-toolbox:42
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "registry.fedoraproject.org/fedora-toolbox:42", decl.Image)
	assert.Empty(t, decl.Distro)
}

func TestParseImageAndDistroMutuallyExclusive(t *testing.T) {
	yamlContent := `
image: fedora:42
distro: fedora
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	_, err = Parse(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestParseWithPkgs(t *testing.T) {
	yamlContent := `
distro: fedora
release: "42"
with-pkgs:
  dnf: [fish, vim]
  apt: [fish, vim]
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	assert.Contains(t, decl.WithPkgs["dnf"], "fish")
	assert.Contains(t, decl.WithPkgs["apt"], "vim")
}

func TestParseMounts(t *testing.T) {
	yamlContent := `
distro: fedora
mounts:
  - source: .
    target: /workspace
    readonly: false
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	require.Len(t, decl.Mounts, 1)
	assert.Equal(t, "/workspace", decl.Mounts[0].Target)
	assert.False(t, decl.Mounts[0].ReadOnly)
}

func TestParseInitHooks(t *testing.T) {
	yamlContent := `
distro: fedora
init-hooks:
  post-create: ["sh", "-c", "echo first"]
  post-start: ["sh", "-c", "echo every"]
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	require.NotNil(t, decl.InitHooks)
	assert.Equal(t, []string{"sh", "-c", "echo first"}, decl.InitHooks.PostCreate)
	assert.Equal(t, []string{"sh", "-c", "echo every"}, decl.InitHooks.PostStart)
}

func TestFindDeclarationWalksUp(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "project")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)

	yamlContent := `distro: fedora`
	err = os.WriteFile(filepath.Join(dir, "toad.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, path, err := Find(subdir)
	require.NoError(t, err)
	assert.Equal(t, "fedora", decl.Distro)
	assert.Equal(t, dir, filepath.Dir(path))
}

func TestFindDeclarationNotFound(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Find(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestParseEnvVars(t *testing.T) {
	yamlContent := `
distro: fedora
env:
  EDITOR: vim
  VISUAL: vim
`
	dir := t.TempDir()
	path := filepath.Join(dir, "toad.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	require.NoError(t, err)

	decl, err := Parse(path)
	require.NoError(t, err)
	assert.Equal(t, "vim", decl.Env["EDITOR"])
	assert.Equal(t, "vim", decl.Env["VISUAL"])
}
