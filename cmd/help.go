package cmd

import (
	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Display help information about Toad",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_ = cmd.Root().Help()
			return nil
		}
		targetCmd, _, err := cmd.Root().Find(args)
		if err != nil {
			return err
		}
		_ = targetCmd.Help()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
