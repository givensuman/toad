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
	"github.com/givensuman/toad/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	upFlags struct {
		path string
	}
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and enter a declarative development container",
	RunE:  up,
}

func init() {
	flags := upCmd.Flags()

	flags.StringVar(&upFlags.path,
		"path",
		"",
		"Path to the directory containing toad.yaml")

	upCmd.SetHelpFunc(upHelp)
	rootCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) error {
	if utils.IsInsideContainer() {
		if !utils.IsInsideToolboxContainer() {
			return errors.New("this is not a Toolbx container")
		}

		exitCode, err := utils.ForwardToHost()
		return &exitError{exitCode, err}
	}

	result, err := declaration.Up(&declaration.UpOptions{
		Path: upFlags.path,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Container '%s' is ready.\n", result.Container)

	return nil
}

func upHelp(cmd *cobra.Command, args []string) {
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
