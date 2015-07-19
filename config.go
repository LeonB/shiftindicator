package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	configFile = "shiftindicator.yml"
)

var (
	ErrMalformedYaml = errors.New("Malformed yaml (%v)")
	ErrUnknownCar = errors.New("Unknown car (%v)")
)

func newConfig() (*config, error) {
	c := &config{}

	configFile, err := c.findConfigFile()
	if err != nil {
		return nil, err
	}

	path, err := os.Open(configFile)
	defer path.Close()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		return nil, fmt.Errorf(ErrMalformedYaml.Error(), err)
	}

	return c, nil
}

type config struct {
	// Shiftpoints []Shiftpoint
	Volume              float64
	Sound               string
	MinTimeBetweenBeeps int `yaml:"min_time_between_beeps"`
	Shiftpoints         map[string][]float32
}

func (c *config) findConfigFile() (string, error) {
	// Find xdg path
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		homedir := user.HomeDir
		xdg = filepath.Join(homedir, ".config")
	}

	// Try xdg path
	path := filepath.Join(xdg, configFile)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try cwd()
	if _, err := os.Stat(configFile); err == nil {
		return configFile, nil
	}

	// Try directory of binary
	binaryDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	path = filepath.Join(binaryDir, configFile)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func (c *config) findSoundFile() (string, error) {
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

func (c *config) getCars() ([]string, error) {
	shiftpoints := c.Shiftpoints
	keys := make([]string, 0, len(shiftpoints))
	for k := range shiftpoints {
		keys = append(keys, k)
	}
	return keys, nil
}

func (c *config) getShiftpointsForCarID(carID string) ([]float32, error) {
	shiftpoints, ok := c.Shiftpoints[carID]
	if ok == false {
		return nil, fmt.Errorf(ErrUnknownCar.Error(), carID)
	}

	if len(shiftpoints) == 0 {
		return nil, fmt.Errorf(ErrNoShiftpoints.Error(), carID)
	}

	return shiftpoints, nil
}
