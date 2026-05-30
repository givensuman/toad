package cmd

import (
	"fmt"

	"github.com/givensuman/toad/pkg/declaration"
	"github.com/givensuman/toad/pkg/podman"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	downFlags struct {
		path string
		rmi  bool
	}
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove a declarative development container",
	RunE:  down,
}

func init() {
	flags := downCmd.Flags()

	flags.StringVarP(&downFlags.path,
		"path",
		"p",
		"",
		"Path to the directory containing toad.yaml")

	flags.BoolVar(&downFlags.rmi,
		"rmi",
		false,
		"Also remove the image")

	rootCmd.AddCommand(downCmd)
}

func down(cmd *cobra.Command, args []string) error {
	if err := requireOutsideContainer(); err != nil {
		return err
	}

	container, err := declaration.Down(&declaration.DownOptions{
		Path: downFlags.path,
		Rmi:  downFlags.rmi,
	})
	if err != nil {
		return err
	}

	var image string
	if downFlags.rmi {
		logrus.Debugf("Inspecting container %s for image", container)
		ctr, err := podman.InspectContainer(container)
		if err != nil {
			logrus.Debugf("Failed to inspect container: %s", err)
		} else {
			image = ctr.Image()
		}
	}

	logrus.Debugf("Removing container %s", container)
	if err := podman.RemoveContainer(container, true); err != nil {
		return err
	}
	fmt.Printf("Removed container: %s\n", container)

	if image != "" {
		logrus.Debugf("Removing image %s", image)
		if err := podman.RemoveImage(image, false); err != nil {
			return fmt.Errorf("failed to remove image %s: %w", image, err)
		}
		fmt.Printf("Removed image: %s\n", image)
	}

	return nil
}
