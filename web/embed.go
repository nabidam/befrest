package web

import "embed"

// Files contains the generated SPA bundled into the Befrest binary.
//
//go:embed all:dist
var Files embed.FS
