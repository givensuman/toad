package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type initYAML struct {
	Distro  string `yaml:"distro"`
	Release string `yaml:"release"`
}

var initFlags struct {
	path string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a starter toad.yaml in the current directory",
	RunE:  initRun,
}

func init() {
	flags := initCmd.Flags()
	flags.StringVarP(&initFlags.path, "path", "p", "", "Directory to create toad.yaml in")

	rootCmd.AddCommand(initCmd)
}

func initRun(cmd *cobra.Command, args []string) error {
	dir := initFlags.path
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	path := filepath.Join(dir, "toad.yaml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	data, err := yaml.Marshal(&initYAML{Distro: "fedora", Release: "42"})
	if err != nil {
		return fmt.Errorf("failed to generate %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	fmt.Printf("Created %s\n", path)
	fmt.Printf("Edit it, then run 'toad up' to create your dev container.\n")
	return nil
}
