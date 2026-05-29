package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func preRunIsCoreOSBug() error {
	return nil
}

func rootRunImpl(cmd *cobra.Command, args []string) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "missing command\n")
	fmt.Fprintf(&builder, "\n")

	usage := getUsageForCommonCommands()
	fmt.Fprintf(&builder, "%s", usage)

	fmt.Fprintf(&builder, "\n")
	fmt.Fprintf(&builder, "Run '%s --help' for usage.", executableBase)

	errMsg := builder.String()
	return errors.New(errMsg)
}
