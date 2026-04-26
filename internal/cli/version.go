package cli

// Build-time variables populated via -ldflags="-X ...". Never set manually.
var (
	// Version is the semver tag of the build (e.g. "v0.1.0"), or "dev"
	// for local builds.
	Version = "dev"

	// Commit is the short git SHA of the build, or "none" for local
	// builds.
	Commit = "none"

	// Date is the RFC 3339 build timestamp, or "unknown" for local
	// builds.
	Date = "unknown"
)
