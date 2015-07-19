package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/codegangsta/cli"
)

const (
	defaultCommand = "run"
	profileFile    = "shiftindicator.prof"
)

func main() {
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	c := cli.NewApp()
	c.Name = "shiftindicator"
	c.Usage = "Shiftindicator for iRacing"
	c.Version = "0.0.1"
	c.Authors = []cli.Author{
		cli.Author{
			Name:  "Leon Bogaert",
			Email: "leonbogaert@gmail.com"},
	}
	c.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "profile",
			Usage: fmt.Sprintf("Enable profiling to %v", profileFile),
		},
	}
	c.Action = func(context *cli.Context) {
		// Do normal stuff
		if context.Args().Present() {
			cli.ShowAppHelp(context)
			return
		}

		// Execute default command
		command := context.App.Command(defaultCommand)
		if c != nil {
			command.Run(context)
		}
	}
	c.Commands = []cli.Command{
		{
			Name:  "shiftpoints",
			Usage: "Shows all defined shiftpoints",
			Flags: nil,
			Action: func(c *cli.Context) {
				cars, _ := app.config.getCars()
				for _, car := range cars {
					shiftpoints, err := app.getShiftpointsForCar(car)
					if err != nil {
						fmt.Println(err)
					}

					fmt.Printf("%s shiftpoints: %+v\n", car, shiftpoints)
				}
			},
		},
		{
			Name:  "play",
			Usage: "Play beep sound",
			Flags: nil,
			Action: func(c *cli.Context) {
				err := app.beep()
				if err != nil {
					log.Fatal(err)
				}
				time.Sleep(app.sound.Total())
			},
		},
		{
			Name:  "run",
			Usage: "Runs the shiftindicator",
			Action: func(c *cli.Context) {
				profile := c.GlobalBool("profile")
				if profile == true {
					f, err := os.Create(profileFile)
					if err != nil {
						log.Fatal(err)
					}

					log.Println("Start profiling")
					pprof.StartCPUProfile(f)
					defer pprof.StopCPUProfile()
				}

				err := app.run()
				if err != nil {
					log.Fatal(err)
				}
			},
		},
	}

	c.Run(os.Args)
}
