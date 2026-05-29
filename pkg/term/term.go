package term

import (
	"os"

	"golang.org/x/sys/unix"
)

type Option func(*unix.Termios)

func GetState(file *os.File) (*unix.Termios, error) {
	fileFD := file.Fd()
	fileFDInt := int(fileFD)
	state, err := unix.IoctlGetTermios(fileFDInt, unix.TCGETS)
	return state, err
}

func IsTerminal(file *os.File) bool {
	if _, err := GetState(file); err != nil {
		return false
	}

	return true
}

func NewStateFrom(oldState *unix.Termios, options ...Option) *unix.Termios {
	newState := *oldState
	for _, option := range options {
		option(&newState)
	}

	return &newState
}

func SetState(file *os.File, state *unix.Termios) error {
	fileFD := file.Fd()
	fileFDInt := int(fileFD)
	err := unix.IoctlSetTermios(fileFDInt, unix.TCSETS, state)
	return err
}

func WithVMIN(vmin uint8) Option {
	return func(state *unix.Termios) {
		state.Cc[unix.VMIN] = vmin
	}
}

func WithVTIME(vtime uint8) Option {
	return func(state *unix.Termios) {
		state.Cc[unix.VTIME] = vtime
	}
}

func WithoutECHO() Option {
	return func(state *unix.Termios) {
		state.Lflag &^= unix.ECHO
	}
}

func WithoutICANON() Option {
	return func(state *unix.Termios) {
		state.Lflag &^= unix.ICANON
	}
}
