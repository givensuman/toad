package declaration

import (
	"fmt"
	"os"

	"github.com/givensuman/go-namesgenerator"
	"github.com/sirupsen/logrus"
)

func (d *Declaration) PackagesForManager(manager string) []string {
	if d.WithPkgs == nil {
		return nil
	}
	return d.WithPkgs[manager]
}

func (d *Declaration) MergePackages(manager string, configPkgs []string) []string {
	declPkgs := d.PackagesForManager(manager)
	seen := make(map[string]struct{}, len(declPkgs)+len(configPkgs))
	result := make([]string, 0, len(declPkgs)+len(configPkgs))

	for _, pkg := range declPkgs {
		if _, ok := seen[pkg]; !ok {
			seen[pkg] = struct{}{}
			result = append(result, pkg)
		}
	}
	for _, pkg := range configPkgs {
		if _, ok := seen[pkg]; !ok {
			seen[pkg] = struct{}{}
			result = append(result, pkg)
		}
	}
	return result
}

func (d *Declaration) EnvVars() []string {
	if d.Env == nil {
		return nil
	}
	vars := make([]string, 0, len(d.Env))
	for k, v := range d.Env {
		vars = append(vars, k+"="+v)
	}
	return vars
}

type UpOptions struct {
	Path string
}

type UpResult struct {
	Container string
}

func Up(opts *UpOptions) (*UpResult, error) {
	workDir := opts.Path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	decl, _, err := Find(workDir)
	if err != nil {
		return nil, err
	}

	container := decl.Container
	if container == "" {
		container = namesgenerator.GetRandomName(0)
	}

	return &UpResult{Container: container}, nil
}

type DownOptions struct {
	Path string
	Rmi  bool
}

func Down(opts *DownOptions) (string, error) {
	workDir := opts.Path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	decl, _, err := Find(workDir)
	if err != nil {
		return "", err
	}

	if decl.Container == "" {
		return "", fmt.Errorf("toad.yaml does not specify a container name; use 'toad rm' to remove by name")
	}

	logrus.Debugf("Resolved container: %s", decl.Container)
	return decl.Container, nil
}
