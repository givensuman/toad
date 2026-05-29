/*
 * Copyright © 2024 – 2026 Red Hat Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/givensuman/toad/pkg/declaration"
	"github.com/givensuman/toad/pkg/podman"
	"github.com/givensuman/toad/pkg/utils"
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

	flags.StringVar(&downFlags.path,
		"path",
		"",
		"Path to the directory containing toad.yaml")

	flags.BoolVar(&downFlags.rmi,
		"rmi",
		false,
		"Also remove the image")

	downCmd.SetHelpFunc(downHelp)
	rootCmd.AddCommand(downCmd)
}

func down(cmd *cobra.Command, args []string) error {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			return errors.New("this is not a Toolbx container")
		}

		exitCode, err := utils.ForwardToHost()
		return &exitError{exitCode, err}
	}

	container, err := declaration.Down(&declaration.DownOptions{
		Path: downFlags.path,
		Rmi:  downFlags.rmi,
	})
	if err != nil {
		return err
	}

	logrus.Debugf("Removing container %s", container)
	if err := podman.RemoveContainer(container, true); err != nil {
		return err
	}
	fmt.Printf("Removed container: %s\n", container)

	if downFlags.rmi {
		logrus.Debugf("Inspecting container %s for image", container)
		ctr, err := podman.InspectContainer(container)
		if err != nil {
			logrus.Debugf("Failed to inspect removed container: %s", err)
			return nil
		}
		image := ctr.Image()
		logrus.Debugf("Removing image %s", image)
		if err := podman.RemoveImage(image, false); err != nil {
			return fmt.Errorf("failed to remove image %s: %w", image, err)
		}
		fmt.Printf("Removed image: %s\n", image)
	}

	return nil
}

func downHelp(cmd *cobra.Command, args []string) {
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

	cmd.Help()
}
