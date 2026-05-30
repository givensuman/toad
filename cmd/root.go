package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/givensuman/toad/pkg/nvidia"
	"github.com/givensuman/toad/pkg/podman"
	"github.com/givensuman/toad/pkg/utils"
	"github.com/givensuman/toad/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cgroupsVersion int

	currentUser *user.User

	executable string

	executableBase string

	rootCmd = &cobra.Command{
		Use:               "toad",
		Short:             "Declarative development containers powered by Podman",
		PersistentPreRunE: preRun,
		RunE:              rootRun,
		Version:           version.GetVersion(),
	}

	rootFlags struct {
		assumeYes bool
		logLevel  string
		logPodman bool
		quiet     bool
		verbose   int
	}

	workingDirectory string
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err != nil {
		return e.err.Error()
	} else {
		return ""
	}
}

func requireOutsideContainer() error {
	if !utils.IsInsideContainer() {
		return nil
	}

	if !utils.IsInsideToolboxContainer() {
		return errors.New("this is not a Toad container")
	}

	exitCode, err := utils.ForwardToHost()
	return &exitError{exitCode, err}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if rootCmd.SilenceErrors {
			if errMsg := err.Error(); errMsg != "" {
				fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
			}
		}

		var errExit *exitError
		if errors.As(err, &errExit) {
			os.Exit(errExit.code)
		}

		os.Exit(1)
	}

	os.Exit(0)
}

func init() {
	if err := setUpGlobals(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.BoolVarP(&rootFlags.assumeYes,
		"assumeyes",
		"y",
		false,
		"Automatically answer yes for all questions")

	persistentFlags.StringVar(&rootFlags.logLevel,
		"log-level",
		"error",
		"Log messages at the specified level: trace, debug, info, warn, error, fatal or panic")

	persistentFlags.BoolVar(&rootFlags.logPodman,
		"log-podman",
		false,
		"Show the log output of Podman. The log level is handled by the log-level option")

	persistentFlags.CountVarP(&rootFlags.verbose, "verbose", "v", "Set log-level to 'debug'")

	persistentFlags.BoolVarP(&rootFlags.quiet, "quiet", "q", false, "Suppress all non-error output")

	logLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	completionFn := cobra.FixedCompletions(logLevels, cobra.ShellCompDirectiveNoFileComp)
	if err := rootCmd.RegisterFlagCompletionFunc("log-level", completionFn); err != nil {
		panicMsg := fmt.Sprintf("failed to register flag completion function: %v", err)
		panic(panicMsg)
	}

	rootCmd.SetHelpTemplate(`Usage:  {{.Use}} [command]

{{.Short}}

Commands:
{{range .VisibleCommands}}{{.Name | printf "  %-12s"}}{{.Short}}
{{end}}

Run '{{.Use}} <command> --help' for more details on a command.
`)
}

func preRun(cmd *cobra.Command, args []string) error {
	cmd.Root().SilenceErrors = true
	cmd.Root().SilenceUsage = true

	if err := setUpLoggers(); err != nil {
		return err
	}

	logrus.Debugf("Running as real user ID %s", currentUser.Uid)
	logrus.Debugf("Resolved absolute path to the executable as %s", executable)

	if !utils.IsInsideContainer() {
		logrus.Debugf("Running on a cgroups v%d host", cgroupsVersion)

	}

	toolbxDelayEntryPoint := os.Getenv("TOOLBX_DELAY_ENTRY_POINT")
	logrus.Debugf("TOOLBX_DELAY_ENTRY_POINT is %s", toolbxDelayEntryPoint)

	toolbxFailEntryPoint := os.Getenv("TOOLBX_FAIL_ENTRY_POINT")
	logrus.Debugf("TOOLBX_FAIL_ENTRY_POINT is %s", toolbxFailEntryPoint)

	toolboxPath := os.Getenv("TOOLBOX_PATH")

	if toolboxPath == "" {
		if utils.IsInsideContainer() {
			if err := preRunIsCoreOSBug(); err != nil {
				return err
			}

			return errors.New("TOOLBOX_PATH not set")
		}

		_ = os.Setenv("TOOLBOX_PATH", executable)
		toolboxPath = os.Getenv("TOOLBOX_PATH")
	}

	logrus.Debugf("TOOLBOX_PATH is %s", toolboxPath)

	if err := migrate(cmd, args); err != nil {
		return err
	}

	if err := utils.SetUpConfiguration(); err != nil {
		return err
	}

	return nil
}

func rootRun(cmd *cobra.Command, args []string) error {
	return rootRunImpl(cmd, args)
}

func migrate(cmd *cobra.Command, args []string) error {
	logrus.Debug("Migrating to newer Podman")

	if utils.IsInsideContainer() {
		logrus.Debug("Migration not needed: running inside a container")
		return nil
	}

	if cmdName, completionCmdName := cmd.Name(), completionCmd.Name(); cmdName == completionCmdName {
		logrus.Debugf("Migration not needed: command %s doesn't need it", cmdName)
		return nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		logrus.Debugf("Migrating to newer Podman: failed to get the user config directory: %s", err)
		return errors.New("failed to get the user config directory")
	}

	toolboxConfigDir := configDir + "/toad"
	stampPath := toolboxConfigDir + "/podman-system-migrate"
	logrus.Debugf("Toad config directory is %s", toolboxConfigDir)

	podmanVersion, err := podman.GetVersion()
	if err != nil {
		logrus.Debugf("Migrating to newer Podman: failed to get the Podman version: %s", err)
		return errors.New("failed to get the Podman version")
	}

	logrus.Debugf("Current Podman version is %s", podmanVersion)

	err = os.MkdirAll(toolboxConfigDir, 0775)
	if err != nil {
		logrus.Debugf("Migrating to newer Podman: failed to create configuration directory %s: %s",
			toolboxConfigDir,
			err)
		return errors.New("failed to create configuration directory")
	}

	toolboxRuntimeDirectory, err := utils.GetRuntimeDirectory(currentUser)
	if err != nil {
		return err
	}

	migrateLock := toolboxRuntimeDirectory + "/migrate.lock"

	migrateLockFile, err := utils.Flock(migrateLock, syscall.LOCK_EX)
	if err != nil {
		logrus.Debugf("Migrating to newer Podman: %s", err)

		var errFlock *utils.FlockError

		if errors.As(err, &errFlock) {
			if errors.Is(err, utils.ErrFlockAcquire) {
				err = utils.ErrFlockAcquire
			} else if errors.Is(err, utils.ErrFlockCreate) {
				err = utils.ErrFlockCreate
			} else {
				panicMsg := fmt.Sprintf("unexpected %T: %s", err, err)
				panic(panicMsg)
			}
		}

		return err
	}

	defer func() { _ = migrateLockFile.Close() }()

	stampBytes, err := os.ReadFile(stampPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Debugf("Migrating to newer Podman: failed to read migration stamp file %s: %s",
				stampPath,
				err)
			return errors.New("failed to read migration stamp file")
		}
	} else {
		stampString := string(stampBytes)
		podmanVersionOld := strings.TrimSpace(stampString)
		if podmanVersionOld != "" {
			logrus.Debugf("Old Podman version is %s", podmanVersionOld)

			if podmanVersion == podmanVersionOld {
				logrus.Debugf("Migration not needed: Podman version %s is unchanged", podmanVersion)
				return nil
			}

			if !podman.CheckVersion(podmanVersionOld) {
				logrus.Debugf("Migration not needed: Podman version %s is old", podmanVersion)
				return nil
			}
		}
	}

	if err = podman.SystemMigrate(""); err != nil {
		logrus.Debugf("Migrating to newer Podman: failed to migrate containers: %s", err)
		return errors.New("failed to migrate containers")
	}

	logrus.Debugf("Migration to Podman version %s was ok", podmanVersion)
	logrus.Debugf("Updating Podman version in %s", stampPath)

	podmanVersionBytes := []byte(podmanVersion + "\n")
	err = os.WriteFile(stampPath, podmanVersionBytes, 0664)
	if err != nil {
		logrus.Debugf("Migrating to newer Podman: failed to update Podman version in migration stamp file %s: %s",
			stampPath,
			err)
		return errors.New("failed to update Podman version in migration stamp file")
	}

	return nil
}

func setUpGlobals() error {
	var err error

	if !utils.IsInsideContainer() {
		cgroupsVersion, err = utils.GetCgroupsVersion()
		if err != nil {
			return fmt.Errorf("failed to get the cgroups version: %w", err)
		}
	}

	currentUser, err = user.Current()
	if err != nil {
		return fmt.Errorf("failed to get the current user: %w", err)
	}

	executable, err = os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get the path to the executable: %w", err)
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path to the executable: %w", err)
	}

	executableBase = filepath.Base(executable)

	workingDirectory, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get the working directory: %w", err)
	}

	return nil
}

func setUpLoggers() error {
	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	if rootFlags.verbose > 0 {
		rootFlags.logLevel = "debug"
	}

	logLevel, err := logrus.ParseLevel(rootFlags.logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log-level: %w", err)
	}

	logrus.SetLevel(logLevel)

	if rootFlags.quiet {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	if rootFlags.verbose > 1 {
		nvidia.SetLogLevel(logLevel)
		rootFlags.logPodman = true
	}

	if rootFlags.logPodman {
		podman.SetLogLevel(logLevel)
	}

	return nil
}
