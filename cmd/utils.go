package cmd

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/givensuman/toad/pkg/shell"
	"github.com/givensuman/toad/pkg/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type askForConfirmationPreFunc func() error
type pollFunc func(error, []unix.PollFd) error

var (
	errClosed = errors.New("closed")

	errContinue = errors.New("continue")

	errHUP = errors.New("HUP")
)

func usageError(format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return errors.New(msg + "\nRun '" + executableBase + " --help' for usage.")
}

// askForConfirmation prints prompt to stdout and waits for response from the
// user
//
// Expected answers are: "yes", "y", "no", "n"
//
// Answers are internally converted to lower case.
//
// The default answer is "no" ([y/N])
func askForConfirmation(prompt string) bool {
	var retVal bool

	ctx := context.Background()
	retValCh, errCh := askForConfirmationAsync(ctx, prompt, nil)

	select {
	case val := <-retValCh:
		retVal = val
	case err := <-errCh:
		logrus.Debugf("Failed to ask for confirmation: %s", err)
		retVal = false
	}

	return retVal
}

func askForConfirmationAsync(ctx context.Context,
	prompt string,
	askForConfirmationPreFn askForConfirmationPreFunc) (<-chan bool, <-chan error) {

	retValCh := make(chan bool, 1)
	errCh := make(chan error, 1)

	done := ctx.Done()
	eventFD := -1
	if done != nil {
		fd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
		if err != nil {
			errCh <- fmt.Errorf("eventfd(2) failed: %w", err)
			return retValCh, errCh
		}

		eventFD = fd
	}

	go func() {
		for {
			fmt.Printf("%s ", prompt)
			if askForConfirmationPreFn != nil {
				if err := askForConfirmationPreFn(); err != nil {
					if errors.Is(err, errContinue) {
						continue
					}

					errCh <- err
					break
				}
			}

			var response string

			pollFn := func(errPoll error, pollFDs []unix.PollFd) error {
				if len(pollFDs) != 1 {
					panic("unexpected number of file descriptors")
				}

				if errPoll != nil {
					return errPoll
				}

				if pollFDs[0].Revents&unix.POLLIN != 0 {
					logrus.Debug("Returned from /dev/stdin: POLLIN")

					scanner := bufio.NewScanner(os.Stdin)
					if !scanner.Scan() {
						if err := scanner.Err(); err != nil {
							return err
						} else {
							return io.EOF
						}
					}

					response = scanner.Text()
					return nil
				}

				if pollFDs[0].Revents&unix.POLLHUP != 0 {
					logrus.Debug("Returned from /dev/stdin: POLLHUP")
					return errHUP
				}

				if pollFDs[0].Revents&unix.POLLNVAL != 0 {
					logrus.Debug("Returned from /dev/stdin: POLLNVAL")
					return errClosed
				}

				return errContinue
			}

			stdinFD := int32(os.Stdin.Fd())

			err := poll(pollFn, int32(eventFD), stdinFD)
			if err != nil {
				errCh <- err
				break
			}

			if response == "" {
				response = "n"
			} else {
				response = strings.ToLower(response)
			}

			if response == "no" || response == "n" {
				retValCh <- false
				break
			} else if response == "yes" || response == "y" {
				retValCh <- true
				break
			}
		}
	}()

	watchContextForEventFD(ctx, eventFD)
	return retValCh, errCh
}

func createErrorContainerNotFound(container string) error {
	return usageError("container %s not found\nUse the 'create' command to create a Toad.", container)
}

func createErrorDistroWithoutRelease(distro string) error {
	return usageError("option '--release' is needed\nDistribution %s doesn't match the host.", distro)
}

func createErrorInvalidContainer(containerArg string) error {
	return usageError("invalid argument for '%s'\nContainer names must match '%s'.", containerArg, utils.ContainerNameRegexp)
}

func createErrorInvalidDistro(distro string) error {
	return usageError("invalid argument for '--distro'\nDistribution %s is unsupported.", distro)
}

func createErrorInvalidImageForContainerName(container string) error {
	return usageError("invalid argument for '--image'\nContainer name %s generated from image is invalid.\nContainer names must match '%s'.", container, utils.ContainerNameRegexp)
}

func createErrorInvalidImageWithoutBasename() error {
	return usageError("invalid argument for '--image'\nImages must have basenames.")
}

func createErrorInvalidRelease(hint string) error {
	return usageError("invalid argument for '--release'\n%s", hint)
}

func createErrorProfileDNotFound() error {
	const profileD = "/etc/profile.d"
	return usageError("directory %s not found in container\nThe shell start-up scripts must include files from %s in\ncontainers.\nGo to https://toad.dev/ for further information.", profileD, profileD)
}

func createErrorSudoersDNotFound() error {
	const sudoersD = "/etc/sudoers.d"
	return usageError("directory %s not found in container\nThe sudoers(5) policy must include files from %s in\ncontainers with /etc/pkcs11/modules and p11-kit-client.so.\nGo to https://toad.dev/ for further information.", sudoersD, sudoersD)
}

func getCDIFileForNvidia(targetUser *user.User) (string, error) {
	toolboxRuntimeDirectory, err := utils.GetRuntimeDirectory(targetUser)
	if err != nil {
		return "", err
	}

	cdiFile := filepath.Join(toolboxRuntimeDirectory, "cdi-nvidia.json")
	return cdiFile, nil
}

func getCurrentUserHomeDir() string {
	if currentUser == nil {
		panic("current user unknown")
	}

	userHomeDir, err := os.UserHomeDir()
	if err == nil {
		return userHomeDir
	}

	logrus.Debugf("Getting the current user's home directory: failed to use os.UserHomeDir(): %s", err)
	logrus.Debug("Using user.Current() instead")

	return currentUser.HomeDir
}

func getCurrentUserShell() (string, error) {
	if currentUser == nil {
		panic("current user unknown")
	}

	if userShell := os.Getenv("SHELL"); userShell != "" {
		return userShell, nil
	}

	logrus.Debug("Getting the current user's login shell: failed to read SHELL")
	logrus.Debug("Using 'getent passwd' instead")

	var stderr strings.Builder
	var stdout strings.Builder

	if err := shell.Run("getent", nil, &stdout, &stderr, "passwd", currentUser.Uid); err != nil {
		errString := stderr.String()
		logrus.Debugf("Getting the current user's login shell failed: %s", errString)
		return "", fmt.Errorf("failed to get the current user's login shell: %w", err)
	}

	output := stdout.String()
	passwdLine := strings.TrimSpace(output)
	if len(passwdLine) == 0 {
		return "", errors.New("failed to get the current user's login shell: no getent(1) output")
	}

	passwdLineParts := strings.Split(passwdLine, ":")
	passwdLinePartsCount := len(passwdLineParts)
	if passwdLinePartsCount != 7 {
		logrus.Debugf("Getting the current user's login shell: failed to parse getent(1) output: %s",
			passwdLine)
		return "", errors.New("failed to get the current user's login shell: invalid getent(1) output")
	}

	return passwdLineParts[passwdLinePartsCount-1], nil
}

func poll(pollFn pollFunc, eventFD int32, fds ...int32) error {
	if len(fds) == 0 {
		panic("file descriptors not specified")
	}

	pollFDs := []unix.PollFd{
		{
			Fd:      eventFD,
			Events:  unix.POLLIN,
			Revents: 0,
		},
	}

	for _, fd := range fds {
		pollFD := unix.PollFd{Fd: fd, Events: unix.POLLIN, Revents: 0}
		pollFDs = append(pollFDs, pollFD)
	}

	for {
		if _, err := unix.Poll(pollFDs, -1); err != nil {
			if errors.Is(err, unix.EINTR) {
				logrus.Debugf("Failed to poll(2): %s: ignoring", err)
				continue
			}

			return fmt.Errorf("poll(2) failed: %w", err)
		}

		var err error

		if pollFDs[0].Revents&unix.POLLIN != 0 {
			logrus.Debug("Returned from eventfd: POLLIN")
			err = context.Canceled

			for {
				buffer := make([]byte, 8)
				if n, err := unix.Read(int(eventFD), buffer); n != len(buffer) || err != nil {
					break
				}
			}
		} else if pollFDs[0].Revents&unix.POLLNVAL != 0 {
			logrus.Debug("Returned from eventfd: POLLNVAL")
			err = context.Canceled
		}

		if err := pollFn(err, pollFDs[1:]); !errors.Is(err, errContinue) {
			return err
		}
	}
}

func resolveContainerAndImageNames(container, containerArg, distroCLI, imageCLI, releaseCLI string) (
	string, string, string, error,
) {
	container, image, release, err := utils.ResolveContainerAndImageNames(container,
		distroCLI,
		imageCLI,
		releaseCLI)

	if err != nil {
		var errContainer *utils.ContainerError
		var errDistro *utils.DistroError
		var errImage *utils.ImageError
		var errParseRelease *utils.ParseReleaseError

		if errors.As(err, &errContainer) {
			if errors.Is(err, utils.ErrContainerNameInvalid) {
				if containerArg == "" {
					panicMsg := fmt.Sprintf("unexpected %T without containerArg: %s", err, err)
					panic(panicMsg)
				}

				err := createErrorInvalidContainer(containerArg)
				return "", "", "", err
			} else if errors.Is(err, utils.ErrContainerNameFromImageInvalid) {
				err := createErrorInvalidImageForContainerName(errContainer.Container)
				return "", "", "", err
			} else {
				panicMsg := fmt.Sprintf("unexpected %T: %s", err, err)
				panic(panicMsg)
			}
		} else if errors.As(err, &errDistro) {
			if errors.Is(err, utils.ErrDistroUnsupported) {
				err := createErrorInvalidDistro(errDistro.Distro)
				return "", "", "", err
			} else if errors.Is(err, utils.ErrDistroWithoutRelease) {
				err := createErrorDistroWithoutRelease(errDistro.Distro)
				return "", "", "", err
			} else {
				panicMsg := fmt.Sprintf("unexpected %T: %s", err, err)
				panic(panicMsg)
			}
		} else if errors.As(err, &errImage) {
			if errors.Is(err, utils.ErrImageWithoutBasename) {
				err := createErrorInvalidImageWithoutBasename()
				return "", "", "", err
			} else {
				panicMsg := fmt.Sprintf("unexpected %T: %s", err, err)
				panic(panicMsg)
			}
		} else if errors.As(err, &errParseRelease) {
			err := createErrorInvalidRelease(errParseRelease.Hint)
			return "", "", "", err
		} else {
			return "", "", "", err
		}
	}

	return container, image, release, nil
}

func watchContextForEventFD(ctx context.Context, eventFD int) {
	done := ctx.Done()
	if done == nil {
		return
	}

	if eventFD < 0 {
		panic("invalid file descriptor for eventfd")
	}

	go func() {
		defer func() { _ = unix.Close(eventFD) }()

		<-done
		buffer := make([]byte, 8)
		binary.PutUvarint(buffer, 1)

		if _, err := unix.Write(eventFD, buffer); err != nil {
			panicMsg := fmt.Sprintf("write(2) to eventfd failed: %s", err)
			panic(panicMsg)
		}
	}()
}
