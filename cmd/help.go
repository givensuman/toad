package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/givensuman/toad/pkg/utils"
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:               "help",
	Short:             "Display help information about Toolbx",
	RunE:              help,
	ValidArgsFunction: completionCommands,
}

func init() {
	helpCmd.SetHelpFunc(helpHelp)
	rootCmd.AddCommand(helpCmd)
}

func help(cmd *cobra.Command, args []string) error {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			return errors.New("this is not a Toolbx container")
		}

		exitCode, err := utils.ForwardToHost()
		return &exitError{exitCode, err}
	}

	if err := helpShowManual(args); err != nil {
		return err
	}

	return nil
}

func helpHelp(cmd *cobra.Command, args []string) {
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

	if err := helpShowManual(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return
	}
}

func helpShowManual(args []string) error {
	var manual string

	if len(args) == 0 {
		manual = "toad"
	} else if args[0] == executableBase {
		manual = "toad"
	} else {
		manual = "toad-" + args[0]
	}

	if err := showManual(manual); err != nil {
		return err
	}

	return nil
}
