package utils

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func getDefaultReleaseFedora() (string, error) {
	release, err := getHostVersionID()
	if err != nil {
		return "", err
	}

	return release, nil
}

func getFullyQualifiedImageFedora(image, release string) string {
	imageFull := "registry.fedoraproject.org/" + image
	return imageFull
}

func getP11KitClientPathsFedora() []string {
	paths := []string{"/usr/lib64/pkcs11/p11-kit-client.so"}
	return paths
}

func parseReleaseFedora(release string) (string, error) {
	if strings.HasPrefix(release, "F") || strings.HasPrefix(release, "f") {
		release = release[1:]
	}

	releaseN, err := strconv.Atoi(release)
	if err != nil {
		logrus.Debugf("Parsing release %s as an integer failed: %s", release, err)
		return "", &ParseReleaseError{"The release must be a positive integer."}
	}

	if releaseN <= 0 {
		return "", &ParseReleaseError{"The release must be a positive integer."}
	}

	return release, nil
}
