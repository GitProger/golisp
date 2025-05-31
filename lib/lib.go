package lib

import "embed"

//go:embed src/*.scm
var StandardLibrary embed.FS
