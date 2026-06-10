package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
	"github.com/pantheon-lab/goctl-openapi/action"
)

var (
	version  = "20220621"
	commands = []*cli.Command{
		{
			Name:   "openapi",
			Usage:  "generates openapi.yaml",
			Action: action.Generator,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "api request address",
				},
				&cli.StringFlag{
					Name:  "basepath",
					Usage: "url request prefix",
				},
				&cli.StringFlag{
					Name:  "filename",
					Usage: "openapi save file name",
				},
				&cli.StringFlag{
					Name:  "schemes",
					Usage: "openapi support schemes: http, https, ws, wss",
				},
				&cli.StringFlag{
					Name:  "api",
					Usage: "api file to parse (standalone mode)",
				},
				&cli.StringFlag{
					Name:  "dir",
					Usage: "output directory (standalone mode)",
				},
			},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Usage = "a plugin of goctl to generate openapi.json"
	app.Version = fmt.Sprintf("%s %s/%s", version, runtime.GOOS, runtime.GOARCH)
	app.Commands = commands
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("goctl-openapi: %+v\n", err)
	}
}
