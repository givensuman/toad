package utils

func getDefaultReleaseArch() (string, error) {
	return "latest", nil
}

func getFullyQualifiedImageArch(image, release string) string {
	imageFull := "quay.io/toolbx/" + image
	return imageFull
}

func getP11KitClientPathsArch() []string {
	paths := []string{"/usr/lib/pkcs11/p11-kit-client.so"}
	return paths
}

func parseReleaseArch(release string) (string, error) {
	if release != "latest" && release != "rolling" && release != "" {
		return "", &ParseReleaseError{"The release must be 'latest'."}
	}

	return "latest", nil
}
