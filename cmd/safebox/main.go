package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkorotkov/safebox/internal/utils"
	"github.com/urfave/cli/v2"
)

const version = "v20210822-1"

func main() {
	vcvout, _, err := utils.LaunchProgram("veracrypt", []string{"--version"}, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to find veracrypt\n")
		os.Exit(1)
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "show the versions of safebox itself and veracrypt found",
	}
	veracryptVersionText := strings.TrimSpace(utils.CaseInsensitiveReplace(string(vcvout), "veracrypt", ""))
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("safebox:   %s\nveracrypt: %s\n", c.App.Version, veracryptVersionText)
	}
	app := &cli.App{
		Name:    "safebox",
		Version: version,
		Usage:   "Makes routine usage of Veracrypt crypto containers more compfortable.",
		Before:  before,
		Commands: []*cli.Command{
			statusCommand(),
			mountCommand(),
			unmountCommand(),
		},
	}
	app.Run(os.Args)
}
