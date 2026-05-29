package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/givensuman/toad/pkg/podman"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rmiFlags struct {
		deleteAll   bool
		forceDelete bool
	}
)

var rmiCmd = &cobra.Command{
	Use:               "rmi",
	Short:             "Remove one or more Toad images",
	RunE:              rmi,
	ValidArgsFunction: completionImageNamesFiltered,
}

func init() {
	flags := rmiCmd.Flags()

	flags.BoolVarP(&rmiFlags.deleteAll, "all", "a", false, "Remove all Toad images")

	flags.BoolVarP(&rmiFlags.forceDelete,
		"force",
		"f",
		false,
		"Force the removal of Toad images that are used by Toad containers")

	rootCmd.AddCommand(rmiCmd)
}

func rmi(cmd *cobra.Command, args []string) error {
	if err := requireOutsideContainer(); err != nil {
		return err
	}

	if rmiFlags.deleteAll {
		logrus.Debug("Getting all images")

		toolboxImages, err := podman.GetImages(false)
		if err != nil {
			logrus.Debugf("Getting all images failed: %s", err)
			return errors.New("failed to get images")
		}

		for toolboxImages.Next() {
			image := toolboxImages.Get()
			imageID := image.ID()
			if err := podman.RemoveImage(imageID, rmiFlags.forceDelete); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				continue
			}
		}
	} else {
		if len(args) == 0 {
			return usageError("missing argument for \"rmi\"")
		}

		for _, image := range args {
			imageObj, err := podman.InspectImage(image)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to inspect image %s\n", image)
				continue
			}

			if !imageObj.IsToolbx() {
				fmt.Fprintf(os.Stderr, "Error: %s is not a Toad image\n", image)
				continue
			}

			if err := podman.RemoveImage(image, rmiFlags.forceDelete); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				continue
			}
		}
	}

	return nil
}


