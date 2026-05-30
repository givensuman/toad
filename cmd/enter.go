package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	enterFlags struct {
		container string
		distro    string
		release   string
	}
)

var enterCmd = &cobra.Command{
	Use:               "enter",
	Short:             "Enter a Toad container for interactive use",
	RunE:              enter,
	ValidArgsFunction: completionContainerNamesFiltered,
}

func init() {
	flags := enterCmd.Flags()

	flags.StringVarP(&enterFlags.container,
		"container",
		"c",
		"",
		"Enter a Toad container with the given name")

	flags.StringVarP(&enterFlags.distro,
		"distro",
		"d",
		"",
		"Enter a Toad container for a different operating system distribution than the host")

	flags.StringVarP(&enterFlags.release,
		"release",
		"r",
		"",
		"Enter a Toad container for a different operating system release than the host")

	if err := enterCmd.RegisterFlagCompletionFunc("container", completionContainerNames); err != nil {
		panicMsg := fmt.Sprintf("failed to register flag completion function: %v", err)
		panic(panicMsg)
	}
	if err := enterCmd.RegisterFlagCompletionFunc("distro", completionDistroNames); err != nil {
		panicMsg := fmt.Sprintf("failed to register flag completion function: %v", err)
		panic(panicMsg)
	}

	rootCmd.AddCommand(enterCmd)
}

func enter(cmd *cobra.Command, args []string) error {
	if err := requireOutsideContainer(); err != nil {
		return err
	}

	var container string
	var containerArg string
	var defaultContainer = true

	if len(args) != 0 {
		container = args[0]
		containerArg = "CONTAINER"
	} else if enterFlags.container != "" {
		container = enterFlags.container
		containerArg = "--container"
	}

	if container != "" {
		defaultContainer = false
	}

	if enterFlags.release != "" {
		defaultContainer = false
	}

	container, image, release, err := resolveContainerAndImageNames(container,
		containerArg,
		enterFlags.distro,
		"",
		enterFlags.release)

	if err != nil {
		return err
	}

	userShell, err := getCurrentUserShell()
	if err != nil {
		return err
	}

	command := []string{userShell, "-l"}

	if err := runCommand(container, defaultContainer, image, release, 0, command, true, true, false); err != nil {
		return err
	}

	return nil
}
