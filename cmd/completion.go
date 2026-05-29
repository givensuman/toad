package cmd

import (
	"os"
	"strings"

	"github.com/givensuman/toad/pkg/podman"
	"github.com/givensuman/toad/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:                   "completion",
	Short:                 "Generate completion script",
	Hidden:                true,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "fish", "zsh"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  completion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func completion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		err := cmd.Root().GenBashCompletionV2(os.Stdout, true)
		return err
	case "fish":
		err := cmd.Root().GenFishCompletion(os.Stdout, true)
		return err
	case "zsh":
		err := cmd.Root().GenZshCompletion(os.Stdout)
		return err
	}

	panic("code should not be reached")
}

func completionEmpty(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completionCommands(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	commandNames := []string{}
	commands := cmd.Root().Commands()
	for _, command := range commands {
		if strings.Contains(command.Name(), "complet") {
			continue
		}
		commandNames = append(commandNames, command.Name())
	}

	return commandNames, cobra.ShellCompDirectiveNoFileComp
}

func completionContainerNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	logrus.Debug("Getting all containers")

	var containerNames []string

	if containers, err := podman.GetContainers(); err != nil {
		logrus.Debugf("Getting all containers failed: %s", err)
	} else {
		for containers.Next() {
			container := containers.Get()
			name := container.Name()
			containerNames = append(containerNames, name)
		}
	}

	return containerNames, cobra.ShellCompDirectiveNoFileComp
}

func completionContainerNamesFiltered(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if cmd.Name() == "enter" && len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	logrus.Debug("Getting all containers")

	var containerNames []string

	if containers, err := podman.GetContainers(); err != nil {
		logrus.Debugf("Getting all containers failed: %s", err)
	} else {
		for containers.Next() {
			container := containers.Get()
			name := container.Name()
			skip := false
			for _, arg := range args {
				if name == arg {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			containerNames = append(containerNames, name)
		}
	}

	return containerNames, cobra.ShellCompDirectiveNoFileComp

}

func completionDistroNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	imageFlag := cmd.Flag("image")
	if imageFlag != nil && imageFlag.Changed {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	supportedDistros := utils.GetSupportedDistros()

	return supportedDistros, cobra.ShellCompDirectiveNoFileComp
}

func completionImageNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	distroFlag := cmd.Flag("distro")
	if distroFlag != nil && distroFlag.Changed {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	logrus.Debug("Getting all images")

	var imageNames []string

	if images, err := podman.GetImages(true); err != nil {
		logrus.Debugf("Getting all images failed: %s", err)
	} else {
		for images.Next() {
			image := images.Get()
			name := image.Name()
			imageNames = append(imageNames, name)
		}
	}

	return imageNames, cobra.ShellCompDirectiveNoFileComp
}

func completionImageNamesFiltered(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	logrus.Debug("Getting all images")

	var imageNames []string

	if images, err := podman.GetImages(true); err != nil {
		logrus.Debugf("Getting all images failed: %s", err)
	} else {
		for images.Next() {
			image := images.Get()
			name := image.Name()
			skip := false
			for _, arg := range args {
				if arg == name {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			imageNames = append(imageNames, name)
		}
	}

	return imageNames, cobra.ShellCompDirectiveNoFileComp
}
