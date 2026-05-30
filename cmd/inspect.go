package cmd

import (
	"fmt"

	"github.com/givensuman/toad/pkg/podman"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:               "inspect",
	Short:             "Display detailed information about a Toad container",
	Args:              cobra.ExactArgs(1),
	RunE:              inspect,
	ValidArgsFunction: completionContainerNamesFiltered,
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}

func inspect(cmd *cobra.Command, args []string) error {
	if err := requireOutsideContainer(); err != nil {
		return err
	}

	container := args[0]
	ctr, err := podman.InspectContainer(container)
	if err != nil {
		return fmt.Errorf("failed to inspect container %s: %w", container, err)
	}

	fmt.Printf("Name:       %s\n", ctr.Name())
	fmt.Printf("ID:         %s\n", ctr.ID())
	fmt.Printf("Image:      %s\n", ctr.Image())
	fmt.Printf("Status:     %s\n", ctr.Status())
	fmt.Printf("Entrypoint: %s\n", ctr.EntryPoint())
	fmt.Printf("Created:    %s\n", ctr.Created())
	fmt.Printf("Labels:     %v\n", ctr.Labels())

	return nil
}
