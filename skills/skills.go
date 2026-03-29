package skills

import "embed"

// FS contains all skill files embedded from the skills directory.
//
//go:embed *.md
var FS embed.FS
