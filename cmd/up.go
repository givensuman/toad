package cmd

import (
	"fmt"

	"github.com/givensuman/toad/pkg/declaration"
	"github.com/spf13/cobra"
)

var (
	upFlags struct {
		path string
	}
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and enter a declarative development container",
	RunE:  up,
}

func init() {
	flags := upCmd.Flags()

	flags.StringVar(&upFlags.path,
		"path",
		"",
		"Path to the directory containing toad.yaml")

	rootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) error {
	if err := requireOutsideContainer(); err != nil {
		return err
	}

	result, err := declaration.Up(&declaration.UpOptions{
		Path: upFlags.path,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Container '%s' is ready.\n", result.Container)

	return nil
}
