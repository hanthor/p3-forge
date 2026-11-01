// Package plugin is the Go SDK for edu plugins. Serve turns a Cobra root
// command into a conforming edu plugin binary.
package plugin

// ManifestSchemaVersion is the schema_version the host writes and reads.
const ManifestSchemaVersion = "1"

// Manifest is what `edu-<name> __metadata` prints and what the host stores
// as edu-plugin.json in the plugin's install directory.
type Manifest struct {
	SchemaVersion string `json:"schema_version"`
	// Name matches the binary suffix: edu-<name>.
	Name    string `json:"name"`
	Version string `json:"version"`
	// Summary is one line shown by `edu help` and `edu plugin list`.
	Summary string `json:"summary"`
	// RequiresHost is a semver range such as ">=0.1.0 <1".
	RequiresHost string `json:"requires_host"`
	// Commands lists leaf commands for help output.
	Commands []Command `json:"commands"`
	Hidden   bool      `json:"hidden,omitempty"`
}

// Command is one leaf command of a plugin, for help output only.
type Command struct {
	// Path is the command words below the plugin name, e.g. ["targets","list"].
	Path    []string `json:"path"`
	Summary string   `json:"summary"`
}
