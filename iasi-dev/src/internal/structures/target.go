package structures

// Target describes one discovered IASI project.
type Target struct {
	Path       string // Project directory.
	Type       string // Literal top-level type from the IASI marker, or "none".
	Depth      int    // Number of discovered ancestor projects.
	Repository string // Containing Git repository, when any.
}
