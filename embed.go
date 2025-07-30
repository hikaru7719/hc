package main

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/out
var frontendFiles embed.FS

func GetFrontendFS() (fs.FS, error) {
	return fs.Sub(frontendFiles, "frontend/out")
}
