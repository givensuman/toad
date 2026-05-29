package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/givensuman/toad/pkg/podman"
	"github.com/givensuman/toad/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rmFlags struct {
		deleteAll   bool
		forceDelete bool
	}
)

var rmCmd = &cobra.Command{
	Use:               "rm",
	Short:             "Remove one or more Toolbx containers",
	RunE:              rm,
	ValidArgsFunction: completionContainerNamesFiltered,
}

func init() {
	flags := rmCmd.Flags()

	flags.BoolVarP(&rmFlags.deleteAll, "all", "a", false, "Remove all Toolbx containers")

	flags.BoolVarP(&rmFlags.forceDelete,
		"force",
		"f",
		false,
		"Force the removal of running and paused Toolbx containers")

	rmCmd.SetHelpFunc(rmHelp)
	rootCmd.AddCommand(rmCmd)
}

func rm(cmd *cobra.Command, args []string) error {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			return errors.New("this is not a Toolbx container")
		}

		exitCode, err := utils.ForwardToHost()
		return &exitError{exitCode, err}
	}

	if rmFlags.deleteAll {
		logrus.Debug("Getting all containers")

		toolboxContainers, err := podman.GetContainers()
		if err != nil {
			logrus.Debugf("Getting all containers failed: %s", err)
			return errors.New("failed to get containers")
		}

		for toolboxContainers.Next() {
			container := toolboxContainers.Get()
			containerID := container.ID()
			if err := podman.RemoveContainer(containerID, rmFlags.forceDelete); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				continue
			}
		}
	} else {
		if len(args) == 0 {
			return usageError("missing argument for \"rm\"")
		}

		for _, container := range args {
			containerObj, err := podman.InspectContainer(container)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to inspect container %s\n", container)
				continue
			}

			if !containerObj.IsToolbx() {
				fmt.Fprintf(os.Stderr, "Error: %s is not a Toolbx container\n", container)
				continue
			}

			if err := podman.RemoveContainer(container, rmFlags.forceDelete); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				continue
			}
		}
	}

	return nil
}

func rmHelp(cmd *cobra.Command, args []string) {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			fmt.Fprintf(os.Stderr, "Error: this is not a Toolbx container\n")
			return
		}

		if _, err := utils.ForwardToHost(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return
		}

		return
	}

	if err := showManual("toolbox-rm"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return
	}
}
