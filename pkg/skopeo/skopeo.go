package skopeo

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/givensuman/toad/pkg/shell"
)

type Layer struct {
	Size json.Number
}
type Image struct {
	LayersData []Layer
}

func Inspect(ctx context.Context, target string) (*Image, error) {
	var stdout bytes.Buffer

	targetWithTransport := "docker://" + target
	args := []string{"inspect", "--format", "json", targetWithTransport}

	if err := shell.RunContext(ctx, "skopeo", nil, &stdout, nil, args...); err != nil {
		return nil, err
	}

	output := stdout.Bytes()
	var image Image
	if err := json.Unmarshal(output, &image); err != nil {
		return nil, err
	}

	return &image, nil
}
