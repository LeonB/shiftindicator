package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
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
						log.Fatal(err)
					}

					fmt.Printf("%s shiftpoints: %+v\n", car, shiftpoints)
				}

				fmt.Println(app.getShiftpointForCarGear("williamsfw31", 4))
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
	}
	c.Run(os.Args)
}
