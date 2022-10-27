package main

import (
	_ "embed"

	"github.com/Gabulhas/polygon-external-consensus/command/root"
	"github.com/Gabulhas/polygon-external-consensus/licenses"
)

var (
	//go:embed LICENSE
	license string
)

func main() {
	licenses.SetLicense(license)

	root.NewRootCommand().Execute()
}
