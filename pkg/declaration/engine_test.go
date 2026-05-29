package declaration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackagesForManager(t *testing.T) {
	d := &Declaration{
		WithPkgs: map[string][]string{
			"dnf": {"fish", "vim"},
			"apt": {"htop"},
		},
	}
	assert.Equal(t, []string{"fish", "vim"}, d.PackagesForManager("dnf"))
	assert.Equal(t, []string{"htop"}, d.PackagesForManager("apt"))
	assert.Nil(t, d.PackagesForManager("apk"))
}

func TestPackagesForManagerNoPkgs(t *testing.T) {
	d := &Declaration{}
	assert.Nil(t, d.PackagesForManager("dnf"))
}

func TestMergePackages(t *testing.T) {
	d := &Declaration{
		WithPkgs: map[string][]string{
			"dnf": {"fish", "vim"},
		},
	}
	result := d.MergePackages("dnf", []string{"starship"})
	assert.ElementsMatch(t, []string{"fish", "vim", "starship"}, result)
}

func TestMergePackagesDeduplicates(t *testing.T) {
	d := &Declaration{
		WithPkgs: map[string][]string{
			"dnf": {"fish", "vim"},
		},
	}
	result := d.MergePackages("dnf", []string{"vim", "starship"})
	assert.ElementsMatch(t, []string{"fish", "vim", "starship"}, result)
}

func TestMergePackagesNoDeclPkgs(t *testing.T) {
	d := &Declaration{}
	result := d.MergePackages("dnf", []string{"starship"})
	assert.Equal(t, []string{"starship"}, result)
}

func TestEnvVars(t *testing.T) {
	d := &Declaration{
		Env: map[string]string{
			"EDITOR": "vim",
			"VISUAL": "vim",
		},
	}
	vars := d.EnvVars()
	assert.ElementsMatch(t, []string{"EDITOR=vim", "VISUAL=vim"}, vars)
}

func TestEnvVarsEmpty(t *testing.T) {
	d := &Declaration{}
	assert.Nil(t, d.EnvVars())
}
