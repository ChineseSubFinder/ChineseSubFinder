package dist

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed spa/index.html
var SpaIndexHtml []byte

//go:embed spa/css
var SpaCSS embed.FS

//go:embed spa/fonts
var SpaFonts embed.FS

//go:embed spa/icons
var SpaIcons embed.FS

//go:embed spa/images
var SpaImages embed.FS

//go:embed spa/js
var SpaJS embed.FS

func Assets(dirName string, emFS embed.FS) http.FileSystem {
	// even uiAssets is empty, fs.Sub won't fail
	stripped, err := fs.Sub(emFS, dirName)
	if err != nil {
		panic(err)
	}
	return http.FS(stripped)
}

const (
	SpaFolderName   = "spa"
	SpaFolderCSS    = "/css"
	SpaFolderFonts  = "/fonts"
	SpaFolderIcons  = "/icons"
	SpaFolderImages = "/images"
	SpaFolderJS     = "/js"
)
