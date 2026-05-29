package declaration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Parse(path string) (*Declaration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var decl Declaration
	if err := yaml.Unmarshal(data, &decl); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if err := validate(&decl); err != nil {
		return nil, err
	}

	return &decl, nil
}

func validate(decl *Declaration) error {
	if decl.Image != "" && decl.Distro != "" {
		return errors.New("image and distro are mutually exclusive")
	}

	if decl.Image == "" && decl.Distro == "" {
		return errors.New("either image or distro must be specified")
	}

	return nil
}

func Find(dir string) (*Declaration, string, error) {
	current := dir
	for {
		candidate := filepath.Join(current, "toad.yaml")
		if _, err := os.Stat(candidate); err == nil {
			decl, err := Parse(candidate)
			if err != nil {
				return nil, "", err
			}
			return decl, candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil, "", fmt.Errorf("toad.yaml not found in %s or any parent directory", dir)
		}
		current = parent
	}
}
