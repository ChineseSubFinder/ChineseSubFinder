package main

import (
	"fmt"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/urfave/cli/v2"
)

/*
	使用的命令：
		chrome_checker -cbp ${chromeBinFPath}

		${chromeBinFPath} -> Chrome 的二进制文件路径

	程序返回值:
		exit(0) success
		exit(1) error
*/

func main() {

	if recover() != nil {
		fmt.Println("Check Chrome has panic")
		os.Exit(1)
	}

	var chromeBinFPath string

	app := &cli.App{
		Name:  "Chrome Checker",
		Usage: "Check Chrome has installed, if not will not auto download and install it",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "chromeBinFPath",
				Aliases:     []string{"cbp"},
				Usage:       "Bin of the browser binary path to launch",
				Destination: &chromeBinFPath,
				Required:    true,
			},
		},
		Action: func(c *cli.Context) error {
			err := rod.Try(func() {
				purl := launcher.New().Bin(chromeBinFPath).MustLaunch()
				browser := rod.New().ControlURL(purl).MustConnect()
				page := browser.MustPage("https://www.baidu.com").MustWaitLoad()
				defer func() {
					_ = page.Close()
				}()
			})
			if err != nil {
				fmt.Println("Check Chrome has error", err.Error())
				os.Exit(1)
			}

			fmt.Println("Check Chrome is ok")
			os.Exit(0)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Check Chrome has error", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
