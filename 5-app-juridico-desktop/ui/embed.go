package ui

import (
	"embed"
	"io/fs"
)

//go:embed static/*
var staticFS embed.FS

// ObterFS retorna o subdiretório static para o file server HTTP embutido no executável
func ObterFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return staticFS
	}
	return sub
}
