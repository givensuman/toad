package version

// currentVersion holds the information about current build version
var (
	currentVersion string
)

// GetVersion returns string with the version of Toolbx
func GetVersion() string {
	return currentVersion
}
