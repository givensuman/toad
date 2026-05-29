package podman

import (
	"fmt"
)

type ImageError struct {
	Image string
	Err   error
}

func (err *ImageError) Error() string {
	errMsg := fmt.Sprintf("%s: %s", err.Image, err.Err)
	return errMsg
}

func (err *ImageError) Unwrap() error {
	return err.Err
}
