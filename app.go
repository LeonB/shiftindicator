package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/leonb/irsdk-go"

	"golang.org/x/mobile/audio"
)

const (
	refreshRateDisconnect = time.Second * 5
	maxFPS                = 30
)

var (
	ErrUnknownGear   = errors.New("Uknown gear")
	ErrNoShiftpoints = errors.New("No shiftpoints found for %s")
	telemetryFields  = []string{"Clutch", "Gear", "RPM"}
)

func newApp() (*App, error) {
	app := &App{
		beepForUpshift:      true,
		gearLastBeepUpshift: 0,
	}

	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	app.config = c

	// Try to load soundfile stored in config
	err = app.loadSound()
	if err != nil {
		return nil, err
	}

	// Initialize connection to iRacing
	app.conn, err = irsdk.NewConnection()
	if err != nil {
		return nil, err
	}
	app.conn.SetMaxFPS(maxFPS)

	return app, nil
}

// App is central struct
type App struct {
	config              *config
	sound               *audio.Player
	conn                *irsdk.Connection
	carID               string
	timeLastBeep        time.Time
	gearLastBeepUpshift int
	beepForUpshift      bool
}

func (a *App) getCarIDFromSession(session *irsdk.SessionData) (string, error) {
	if a.carID == "" {
		driverCar := session.DriverInfo.DriverCarIdx
		for _, driver := range session.DriverInfo.Drivers {
			if driver.CarIdx == driverCar {
				return driver.CarPath, nil
			}
		}

	}
	return "", nil
}

func (a *App) getShiftpointsForCar(carID string) ([]float32, error) {
	shiftpoints, err := a.config.getShiftpointsForCarID(carID)
	if err != nil {
		return nil, err
	}

	return shiftpoints, nil
}

func (a *App) getDefaultShiftpointForCar(carID string) (float32, error) {
	// Do something with irsdk
	return 0, nil
}

func (a *App) getShiftpointForCarGear(carID string, gear int) (float32, error) {
	if gear < 0 {
		return 0, ErrUnknownGear
	}

	if gear < 1 {
		return 0, nil
	}

	shiftpoints, err := a.config.getShiftpointsForCarID(carID)
	if err != nil {
		return 0, err
	}

	if len(shiftpoints) == 1 {
		return shiftpoints[0], nil
	}

	if (gear - 1) > len(shiftpoints) {
		return shiftpoints[len(shiftpoints)-1], ErrUnknownGear
	}

	return shiftpoints[gear-1], nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error {
	return nil
}

func (a *App) beep() error {
	if a.sound == nil {
	}

	err := a.sound.Play()
	if err != nil {
		return err
	}

	a.timeLastBeep = time.Now()
	return nil
}

func (a *App) loadSound() error {
	// Load bytes from soundfile
	soundFile, err := a.config.findSoundFile()
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(soundFile)
	if err != nil {
		return err
	}

	// Add close() method
	buf := &nopCloser{bytes.NewReader(b)}

	// Setup player
	format := audio.Format(0)
	samplesPerSecond := int64(0)
	sound, err := audio.NewPlayer(buf, format, samplesPerSecond)
	if err != nil {
		return err
	}

	sound.SetVolume(a.config.Volume)

	// Store player in app struct
	a.sound = sound

	return nil
}

func (a *App) run() error {
	var err error
	prevConnStatus := false
	curConnStatus := false
	a.conn.Connect()

	for {
		curConnStatus = a.conn.IsConnected()
		if curConnStatus == false {
			if prevConnStatus == true {
				err = a.onSessionEnd()
				if err != nil {
					return err
				}
			}

			// Wait 5 seconds before next connect attempt
			time.Sleep(refreshRateDisconnect)
			a.conn.Connect()
		} else {
			// GetSessionData() blocks until new data is ready
			session, err := a.conn.GetSessionData()
			if err != nil {
				return err
			}

			telemetry, err := a.conn.GetTelemetryDataFiltered(telemetryFields)
			if err != nil {
				prevConnStatus = curConnStatus
				continue
			}

			if prevConnStatus == false {
				err = a.onSessionStart(session, telemetry)
				if err != nil {
					return err
				}
			} else {
				err := a.onTick(session, telemetry)
				if err != nil {
					return err
				}
			}
		}
		// Update connection status
		prevConnStatus = curConnStatus
	}

	return nil
}

func (a *App) onSessionStart(session *irsdk.SessionData, telemetry *irsdk.TelemetryData) error {
	log.Println("onSessionStart")
	var err error

	a.carID, err = a.getCarIDFromSession(session)
	if err != nil {
		return err
	}
	log.Println("car:", a.carID)

	return nil
}

func (a *App) onTick(session *irsdk.SessionData, telemetry *irsdk.TelemetryData) error {
	gear := telemetry.Gear
	rpm := telemetry.RPM
	clutch := telemetry.Clutch

	// Don't beep when clutch is pressed
	if clutch == 0 {
		return nil
	}

	// Don't beep in neutral or reverse gear
	if gear < 1 {
		return nil
	}

	// Get shiftpoint for this car and gear
	shiftpoint, err := a.getShiftpointForCarGear(a.carID, gear)
	if err != nil {
		// Probably unknown car, fetch default shiftpoint for car from
		// sessiondata
		shiftpoint = session.DriverInfo.DriverCarSLShiftRPM
	}

	// Reset beep if a gear has changed or rpm dropped below shiftpoint
	if a.beepForUpshift == false {
		// @TODO: take time into account?
		if rpm < shiftpoint {
			log.Println("rpm below shiftpoint: reset beepForUpshift")
			a.beepForUpshift = true
		} else if gear != a.gearLastBeepUpshift {
			log.Println("changed gear: reset beepForUpshift")
			a.beepForUpshift = true
		}
	}

	// Already beeped this gear: skip beeping
	if a.beepForUpshift == false {
		return nil
	}

	// RPM's are below shiftpoint: don't beep
	if rpm < shiftpoint {
		return nil
	}

	// Don't beep to often: for example if you just dip slightly below the
	// shiftpoint
	duration := time.Now().Sub(a.timeLastBeep)
	time := (time.Millisecond * time.Duration(a.config.MinTimeBetweenBeeps))
	if duration <= time {
		return nil
	}

	// Everything checks out: beep and update data
	err = a.beep()
	if err != nil {
		return err
	}
	a.gearLastBeepUpshift = gear
	a.beepForUpshift = false

	return nil
}

func (a *App) onSessionEnd() error {
	log.Println("onSessionEnd")
	// Reset struct values
	a.carID = ""
	a.timeLastBeep = time.Time{}
	return nil
}
