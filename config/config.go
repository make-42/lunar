package config

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v2"
)

type ConfigS struct {
	Lat                float64
	Lon                float64
	ImageSize          int
	MoonRad            float64
	LineWidth          float64
	LineRes            int
	HorizonLine        bool
	HorizonLineWidth   float64
	HorizonLinePadding float64
	Meridians          bool
	MeridianCount      int
	Parallels          bool
	ParallelCount      int
	GeodesicsThickness float64
	Color              string
}

var DefaultConfig = ConfigS{
	Lat:                48.8,
	Lon:                2.3,
	ImageSize:          1024,
	MoonRad:            400,
	LineWidth:          10.,
	LineRes:            100,
	HorizonLine:        true,
	HorizonLineWidth:   3.,
	HorizonLinePadding: 10.,
	Meridians:          false,
	MeridianCount:      7,
	Parallels:          false,
	ParallelCount:      7,
	GeodesicsThickness: 3.,
	Color:              "#FFFFFF",
}

var Config = ConfigS{}

var OutColor = color.RGBA{}

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}

func Init() {
	configPath := configdir.LocalConfig("ontake", "lunar")
	err := configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		log.Fatal(err)
	}

	configFile := filepath.Join(configPath, "config.yml")

	// Does the file not exist?
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		// Create the new config file.
		fh, err := os.Create(configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()

		encoder := yaml.NewEncoder(fh)
		encoder.Encode(&DefaultConfig)
		Config = DefaultConfig
	} else {
		Config = DefaultConfig
		// Load the existing file.
		fh, err := os.Open(configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()

		decoder := yaml.NewDecoder(fh)
		decoder.Decode(&Config)
	}
	OutColor, err = ParseHexColor(Config.Color)
	if err != nil {
		log.Fatal(err)
	}
}
