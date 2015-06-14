package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/mobile/audio"
	"gopkg.in/yaml.v2"
)

const (
	configFile = "shiftindicator.yml"
)

var (
	ErrUnknownGear   = errors.New("Uknown gear")
	ErrMalformedYaml = errors.New("Malformed yaml (%v)")
	ErrNoShiftpoints = errors.New("No shiftpoints found")
)

func newApp() (*App, error) {
	app := &App{}
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	app.config = c

	// Initialize connection to iRacing

	return app, nil
}

func newConfig() (*Config, error) {
	c := &Config{}

	configFile, err := c.findConfigFile()
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("%s", ErrMalformedYaml), err)
	}

	return c, nil
}

type Config struct {
	// Shiftpoints []Shiftpoint
	Volume               float64
	Sound                string
	MineTimeBetweenBeeps int
	Shiftpoints          map[string][]int
}

func (c *Config) findConfigFile() (*os.File, error) {
	// Find xdg path
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		user, err := user.Current()
		if err != nil {
			return nil, err
		}
		homedir := user.HomeDir
		xdg = filepath.Join(homedir, ".config")
	}

	// Try xdg path
	path := filepath.Join(xdg, configFile)
	if _, err := os.Stat(path); err == nil {
		return os.Open(path)
	}

	// Try cwd()
	if _, err := os.Stat(configFile); err == nil {
		return os.Open(configFile)
	}

	// Try directory of binary
	binaryDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	path = filepath.Join(binaryDir, configFile)
	return os.Open(path)
}

func (c *Config) findSoundFile() (string, error) {
	// Try directory of binary
	binaryDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	path := filepath.Join(binaryDir, c.Sound)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func (c *Config) getCars() ([]string, error) {
	shiftpoints := c.Shiftpoints
	keys := make([]string, 0, len(shiftpoints))
	for k := range shiftpoints {
		keys = append(keys, k)
	}
	return keys, nil
}

func (c *Config) getShiftpointsForCarId(carId string) ([]int, error) {
	shiftpoints, ok := c.Shiftpoints[carId]
	if ok == false {
		return nil, fmt.Errorf("Unknown car")
	}

	if len(shiftpoints) == 0 {
		return nil, ErrNoShiftpoints
	}

	return shiftpoints, nil
}

type App struct {
	config *Config
	sound  *audio.Player
}

func (a *App) getShiftpointsForCar(carId string) ([]int, error) {
	shiftpoints, err := a.config.getShiftpointsForCarId(carId)
	if err != nil {
		return nil, err
	}

	return shiftpoints, nil
}

func (a *App) getDefaultShiftpointForCar(carId string) (int, error) {
	// Do something with irsdk
	return 0, nil
}

func (a *App) getShiftpointForCarGear(carId string, gear int) (int, error) {
	if gear < 0 {
		return 0, ErrUnknownGear
	}

	if gear < 2 {
		return 0, nil
	}

	shiftpoints, err := a.config.getShiftpointsForCarId(carId)
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
		err := a.loadSound()
		if err != nil {
			return err
		}
	}

	err := a.sound.Play()
	if err != nil {
		return err
	}

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

	// Store player in app struct
	a.sound = sound

	return nil
}
