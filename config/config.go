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

type ColorsS struct {
	Default              color.RGBA
	CivilTwilight        color.RGBA
	NauticalTwilight     color.RGBA
	AstronomicalTwilight color.RGBA
	GoldenHour           color.RGBA
	Day                  color.RGBA
	Night                color.RGBA
}

type ColorStrings struct {
	Default              string
	CivilTwilight        string
	NauticalTwilight     string
	AstronomicalTwilight string
	GoldenHour           string
	Day                  string
	Night                string
}

type SolarS struct {
	Width             int
	Height            int
	LineHeight        float64
	LineWidth         float64
	MarkerLineWidth   float64
	MarkerCircleRadius float64
	SunPathLineWidth  float64
	TimeTickLineWidth float64
	TimeTickPeriod    int // hours
	LineRes           int
}

type LunarS struct {
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
}

type ConfigS struct {
	Lat    float64
	Lon    float64
	Lunar  LunarS
	Solar  SolarS
	Colors ColorStrings
}

var DefaultConfig = ConfigS{
	Lat: 48.8,
	Lon: 2.3,
	Lunar: LunarS{
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
	},
	Solar: SolarS{
		Width:             1024,
		Height:            256,
		LineHeight:        200,
		LineWidth:         6.,
		MarkerLineWidth:   4.,
		MarkerCircleRadius: 8.,
		SunPathLineWidth:  10.,
		TimeTickLineWidth: 2.,
		TimeTickPeriod:    2, // hours
		LineRes:           150,
	},
	Colors: ColorStrings{
		Default:              "#E6F0FF", // soft neutral fallback
		Day:                  "#87CEEB", // clear sky blue
		GoldenHour:           "#FFB347", // warm amber
		CivilTwilight:        "#FF8C42", // orange-red glow
		NauticalTwilight:     "#2F4F8F", // deep blue
		AstronomicalTwilight: "#0B1D3A", // near-night blue
		Night:                "#020617", // almost black, but not harsh
	},
}

var Config = ConfigS{}

var ParsedColors = ColorsS{}

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
	parseAllColors()
}

func parseAllColors() {
	cs := Config.Colors

	var err error
	ParsedColors.Default, err = ParseHexColor(cs.Default)
	must(err)
	ParsedColors.CivilTwilight, err = ParseHexColor(cs.CivilTwilight)
	must(err)
	ParsedColors.NauticalTwilight, err = ParseHexColor(cs.NauticalTwilight)
	must(err)
	ParsedColors.AstronomicalTwilight, err = ParseHexColor(cs.AstronomicalTwilight)
	must(err)
	ParsedColors.GoldenHour, err = ParseHexColor(cs.GoldenHour)
	must(err)
	ParsedColors.Day, err = ParseHexColor(cs.Day)
	must(err)
	ParsedColors.Night, err = ParseHexColor(cs.Night)
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
