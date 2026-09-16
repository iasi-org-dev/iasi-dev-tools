package structures

// Target describes one discovered IASI project.
type Target struct {
	Path       string // Project directory containing the IASI marker.
	Type       string // Literal top-level type from the IASI marker, or "none".
	Builder    string // Literal builder from the IASI marker.
	SourceDir  string // Source directory relative to Path when not absolute.
	OutputDir  string // Output directory relative to Path when not absolute.
	Name       string // Artifact name.
	Depth      int    // Number of discovered ancestor projects.
	Repository string // Containing Git repository, when any.
}
