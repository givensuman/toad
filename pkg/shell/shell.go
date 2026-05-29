package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Run(name string, stdin io.Reader, stdout, stderr io.Writer, arg ...string) error {
	ctx := context.Background()
	err := RunContext(ctx, name, stdin, stdout, stderr, arg...)
	return err
}

func RunContext(ctx context.Context, name string, stdin io.Reader, stdout, stderr io.Writer, arg ...string) error {
	exitCode, err := RunContextWithExitCode(ctx, name, stdin, stdout, stderr, arg...)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("failed to invoke %s(1)", name)
	}
	return nil
}

func RunContextWithExitCode(ctx context.Context,
	name string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	arg ...string) (int, error) {

	logLevel := logrus.GetLevel()
	if stderr == nil && logLevel >= logrus.DebugLevel {
		stderr = os.Stderr
	}

	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return 1, fmt.Errorf("%s(1) not found", name)
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return 1, ctxErr
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			return exitCode, nil
		}

		return 1, fmt.Errorf("failed to invoke %s(1)", name)
	}

	return 0, nil
}

func RunWithExitCode(name string, stdin io.Reader, stdout, stderr io.Writer, arg ...string) (int, error) {
	ctx := context.Background()
	exitCode, err := RunContextWithExitCode(ctx, name, stdin, stdout, stderr, arg...)
	return exitCode, err
}
