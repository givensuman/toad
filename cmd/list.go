package cmd

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/givensuman/toad/pkg/podman"
	"github.com/givensuman/toad/pkg/term"
	"github.com/givensuman/toad/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listFlags struct {
		onlyContainers bool
		onlyImages     bool
	}
)

var listCmd = &cobra.Command{
	Use:               "list",
	Short:             "List existing Toad containers and images",
	RunE:              list,
	ValidArgsFunction: completionEmpty,
}

func init() {
	flags := listCmd.Flags()

	flags.BoolVarP(&listFlags.onlyContainers,
		"containers",
		"c",
		false,
		"List only Toad containers, not images")

	flags.BoolVarP(&listFlags.onlyImages,
		"images",
		"i",
		false,
		"List only Toad images, not containers")

	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) error {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			return errors.New("this is not a Toad container")
		}

		exitCode, err := utils.ForwardToHost()
		return &exitError{exitCode, err}
	}

	lsContainers := true
	lsImages := true

	if !listFlags.onlyContainers && listFlags.onlyImages {
		lsContainers = false
	} else if listFlags.onlyContainers && !listFlags.onlyImages {
		lsImages = false
	}

	var images *podman.Images
	var containers *podman.Containers
	var err error

	if lsImages {
		logrus.Debug("Getting all images")

		images, err = podman.GetImages(false)
		if err != nil {
			logrus.Debugf("Getting all images failed: %s", err)
			return errors.New("failed to get images")
		}
	}

	if lsContainers {
		logrus.Debug("Getting all containers")

		containers, err = podman.GetContainers()
		if err != nil {
			logrus.Debugf("Getting all containers failed: %s", err)
			return errors.New("failed to get containers")
		}
	}

	listOutput(images, containers)
	return nil
}

func listOutput(images *podman.Images, containers *podman.Containers) {
	if images.Len() != 0 {
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(writer, "%s\t%s\t%s\n", "IMAGE ID", "IMAGE NAME", "CREATED")

		for images.Next() {
			image := images.Get()
			created := image.Created()
			name := image.Name()

			id := image.ID()
			shortID := utils.ShortID(id)

			fmt.Fprintf(writer, "%s\t%s\t%s\n", shortID, name, created)
		}

		writer.Flush()
	}

	if images.Len() != 0 && containers.Len() != 0 {
		fmt.Println()
	}

	if containers.Len() != 0 {
		const boldGreenColor = "\033[1;32m"
		const defaultColor = "\033[0;00m" // identical to resetColor, but same length as boldGreenColor
		const resetColor = "\033[0m"

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if term.IsTerminal(os.Stdout) {
			fmt.Fprintf(writer, "%s", defaultColor)
		}

		fmt.Fprintf(writer,
			"%s\t%s\t%s\t%s\t%s",
			"CONTAINER ID",
			"CONTAINER NAME",
			"CREATED",
			"STATUS",
			"IMAGE NAME")

		if term.IsTerminal(os.Stdout) {
			fmt.Fprintf(writer, "%s", resetColor)
		}

		fmt.Fprintf(writer, "\n")

		for containers.Next() {
			container := containers.Get()

			isRunning := false
			if podman.CheckVersion("2.0.0") {
				status := container.Status()
				isRunning = status == "running"
			}

			if term.IsTerminal(os.Stdout) {
				var color string
				if isRunning {
					color = boldGreenColor
				} else {
					color = defaultColor
				}

				fmt.Fprintf(writer, "%s", color)
			}

			created := container.Created()
			image := container.Image()
			name := container.Name()

			id := container.ID()
			shortID := utils.ShortID(id)

			status := container.Status()

			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s", shortID, name, created, status, image)

			if term.IsTerminal(os.Stdout) {
				fmt.Fprintf(writer, "%s", resetColor)
			}

			fmt.Fprintf(writer, "\n")
		}

		writer.Flush()
	}
}
