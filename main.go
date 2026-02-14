package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"sort"
	"time"

	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v3"
	"github.com/sixdouglas/suncalc"

	"lunar/config"
)

const moonRadius = 1737.4 // km

func screenCoords(x, y, visAngle float64) (float64, float64) {
	return float64(config.Config.Lunar.ImageSize)/2 + config.Config.Lunar.MoonRad*(math.Cos(visAngle)*x-math.Sin(visAngle)*y), float64(config.Config.Lunar.ImageSize)/2 + config.Config.Lunar.MoonRad*(math.Sin(visAngle)*x+math.Cos(visAngle)*y)
}

func SignedAngleDiff(a, b float64) float64 {
	diff := math.Mod(b-a+math.Pi, 2*math.Pi)
	if diff < 0 {
		diff += 2 * math.Pi
	}
	return diff - math.Pi
}

func genLunarImage(mp suncalc.MoonPosition, mi suncalc.MoonIllumination) *gg.Context {
	visAngle := SignedAngleDiff(mi.Angle, mp.ParallacticAngle)
	//fmt.Println(mp.ParallacticAngle, mi.Angle)
	anglePhase := mi.Phase * (2 * math.Pi)
	//fmt.Println(visAngle*360/(2*math.Pi), mi.Phase)
	dc := gg.NewContext(config.Config.Lunar.ImageSize, config.Config.Lunar.ImageSize)
	dc.SetLineWidth(config.Config.Lunar.LineWidth)
	dc.DrawCircle(float64(config.Config.Lunar.ImageSize)/2, float64(config.Config.Lunar.ImageSize)/2, config.Config.Lunar.MoonRad)
	dc.SetColor(config.ParsedColors.Default)
	dc.Stroke()
	if config.Config.Lunar.Parallels || config.Config.Lunar.Meridians {
		maskdc := gg.NewContext(config.Config.Lunar.ImageSize, config.Config.Lunar.ImageSize)
		maskdc.SetLineWidth(config.Config.Lunar.GeodesicsThickness)

		if config.Config.Lunar.Parallels {
			for i := 0; i <= config.Config.Lunar.ParallelCount; i++ {
				phi := math.Pi * float64(i+1) / float64(config.Config.Lunar.ParallelCount+2)
				currX, currY := screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(0), visAngle)
				maskdc.MoveTo(currX, currY)
				for j := 0; j < config.Config.Lunar.LineRes; j++ {
					theta := math.Pi * float64(j+1) / float64(config.Config.Lunar.LineRes)
					currX, currY = screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(theta), visAngle)
					maskdc.LineTo(currX, currY)
				}
				maskdc.Stroke()
			}
		}
		if config.Config.Lunar.Meridians {
			for i := 0; i <= config.Config.Lunar.MeridianCount; i++ {
				theta := math.Pi * float64(i+1) / float64(config.Config.Lunar.MeridianCount+2)
				currX, currY := screenCoords(math.Cos(0), math.Sin(0)*math.Cos(theta), visAngle)
				maskdc.MoveTo(currX, currY)
				for j := 0; j < config.Config.Lunar.LineRes; j++ {
					phi := math.Pi * float64(j+1) / float64(config.Config.Lunar.LineRes)
					currX, currY = screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(theta), visAngle)
					maskdc.LineTo(currX, currY)
				}
				maskdc.Stroke()
			}
		}

		maskimg := maskdc.AsMask()
		dc.SetMask(maskimg)
		dc.InvertMask()
	}

	currX, currY := screenCoords(1, 0, visAngle)
	dc.MoveTo(currX, currY)
	for i := 0; i < config.Config.Lunar.LineRes; i++ {
		phi := math.Pi * float64(i+1) / float64(config.Config.Lunar.LineRes)
		if math.Sin(anglePhase) > 0 {
			currX, currY = screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(anglePhase), visAngle) // Should be opposite if sin < 0 for anglePhase
		} else {
			currX, currY = screenCoords(math.Cos(phi), -math.Sin(phi)*math.Cos(anglePhase), visAngle) // Should be opposite if sin < 0 for anglePhase
		}
		dc.LineTo(currX, currY)
	}
	for i := 0; i < config.Config.Lunar.LineRes; i++ {
		phi := math.Pi * (1 - float64(i+1)/float64(config.Config.Lunar.LineRes))
		if math.Sin(anglePhase) > 0 {
			currX, currY = screenCoords(math.Cos(phi), math.Sin(phi), visAngle) // Should be opposite if sin < 0 for anglePhase
		} else {
			currX, currY = screenCoords(math.Cos(phi), -math.Sin(phi), visAngle) // Should be opposite if sin < 0 for anglePhase
		}
		dc.LineTo(currX, currY)
	}
	dc.Fill()

	dc.SetLineWidth(config.Config.Lunar.HorizonLineWidth)
	dc.SetDash(5, 10)
	frac := mp.Altitude / (moonRadius / mp.Distance / config.Config.Lunar.MoonRad * float64(config.Config.Lunar.ImageSize))
	if (frac <= 1) && mp.Altitude >= 0 {
		partFrac := (frac*float64(config.Config.Lunar.ImageSize) - (float64(config.Config.Lunar.ImageSize)/2 - config.Config.Lunar.MoonRad)) / (2 * config.Config.Lunar.MoonRad)
		if partFrac >= 0 && partFrac <= 1 {
			// two lines
			y := math.Sqrt(1. - (partFrac-1./2.)*(partFrac-1./2.)*4.)
			dc.DrawLine(config.Config.Lunar.HorizonLinePadding, float64(config.Config.Lunar.ImageSize)*(1-frac), float64(config.Config.Lunar.ImageSize/2)-y*config.Config.Lunar.MoonRad, float64(config.Config.Lunar.ImageSize)*(1-frac))
			dc.DrawLine(float64(config.Config.Lunar.ImageSize)-config.Config.Lunar.HorizonLinePadding, float64(config.Config.Lunar.ImageSize)*(1-frac), float64(config.Config.Lunar.ImageSize/2)+y*config.Config.Lunar.MoonRad, float64(config.Config.Lunar.ImageSize)*(1-frac))
		} else {
			// single line
			dc.DrawLine(config.Config.Lunar.HorizonLinePadding, float64(config.Config.Lunar.ImageSize)*(1-frac), float64(config.Config.Lunar.ImageSize)-config.Config.Lunar.HorizonLinePadding, float64(config.Config.Lunar.ImageSize)*(1-frac))
		}
	}
	dc.Stroke()
	return dc
}

func sunStateColor(currTime time.Time, stl []TimePair) color.RGBA {
	for i := 0; i < len(stl)-1; i++ {
		if stl[i].Time.Before(currTime) && stl[i+1].Time.After(currTime) {
			if stl[i].Name == suncalc.NauticalDawn {
				return config.ParsedColors.NauticalTwilight
			}
			if stl[i].Name == suncalc.Dawn {
				return config.ParsedColors.CivilTwilight
			}
			if stl[i].Name == suncalc.Sunrise || stl[i].Name == suncalc.SunriseEnd {
				return config.ParsedColors.GoldenHour
			}
			if stl[i].Name == suncalc.GoldenHourEnd || stl[i].Name == suncalc.SolarNoon {
				return config.ParsedColors.Day
			}
			if stl[i].Name == suncalc.GoldenHour {
				return config.ParsedColors.GoldenHour
			}
			if stl[i].Name == suncalc.SunsetStart || stl[i].Name == suncalc.Sunset {
				return config.ParsedColors.CivilTwilight
			}
			if stl[i].Name == suncalc.Dusk {
				return config.ParsedColors.NauticalTwilight
			}
			if stl[i].Name == suncalc.NauticalDusk {
				return config.ParsedColors.AstronomicalTwilight
			}
			if stl[i].Name == suncalc.Night || stl[i].Name == suncalc.Nadir {
				return config.ParsedColors.Night
			}
			if stl[i].Name == suncalc.NightEnd {
				return config.ParsedColors.NauticalTwilight
			}
		}
	}
	return config.ParsedColors.Default
}

func refTime(sta map[suncalc.DayTimeName]suncalc.DayTime) time.Time {
	return sta[suncalc.SunsetStart].Value.Add((sta[suncalc.Sunset].Value.Sub(sta[suncalc.SunsetStart].Value)) / 2)
}

func timeToPos(sta map[suncalc.DayTimeName]suncalc.DayTime, t time.Time) float64 {
	ref := refTime(sta)
	return math.Mod((float64(config.Config.Solar.Width)*float64(t.Sub(ref).Seconds())/float64(24*60*60))+float64(config.Config.Solar.Width)/2, float64(config.Config.Solar.Width))
}

func altToPos(minAlt, maxAlt, currAlt float64) float64 {
	return -(float64(config.Config.Solar.LineHeight) * (currAlt - minAlt) / (maxAlt - minAlt)) + float64(config.Config.Solar.LineHeight)/2 + float64(config.Config.Solar.Height)/2
}

func genSolarImage(currTime time.Time) *gg.Context {
	sty := suncalc.GetTimes(currTime.Add(-time.Hour*24), config.Config.Lat, config.Config.Lon)
	sta := suncalc.GetTimes(currTime, config.Config.Lat, config.Config.Lon)
	stb := suncalc.GetTimes(currTime.Add(time.Hour*24), config.Config.Lat, config.Config.Lon)
	stz := suncalc.GetTimes(currTime.Add(time.Hour*48), config.Config.Lat, config.Config.Lon)
	stl := []TimePair{}

	for key := range sty {
		stl = append(stl, TimePair{Time: sty[key].Value, Name: key})
	}
	for key := range sta {
		stl = append(stl, TimePair{Time: sta[key].Value, Name: key})
	}
	for key := range stb {
		stl = append(stl, TimePair{Time: stb[key].Value, Name: key})
	}
	for key := range stz {
		stl = append(stl, TimePair{Time: stz[key].Value, Name: key})
	}
	sort.Slice(stl, func(i, j int) bool {
		return stl[i].Time.Before(stl[j].Time)
	})

	sps := make([]suncalc.SunPosition, config.Config.Solar.LineRes)
	minAlt := math.Pi / 2
	maxAlt := -math.Pi / 2
	spdates := make([]time.Time, config.Config.Solar.LineRes)
	for i := 0; i < config.Config.Solar.LineRes; i++ {
		spdates[i] = currTime.Add(time.Duration(float64(time.Hour*24) * float64(i) / float64(config.Config.Solar.LineRes)))
		sps[i] = suncalc.GetPosition(spdates[i], config.Config.Lat, config.Config.Lon)
		if sps[i].Altitude > maxAlt {
			maxAlt = sps[i].Altitude
		}
		if sps[i].Altitude < minAlt {
			minAlt = sps[i].Altitude
		}
	}

	centerY := altToPos(minAlt, maxAlt, 0)

	dc := gg.NewContext(config.Config.Solar.Width, config.Config.Solar.Height)
	dc.SetLineWidth(config.Config.Solar.LineWidth)
	dc.SetColor(config.ParsedColors.Default)
	dc.DrawLine(0, altToPos(minAlt, maxAlt, 0), float64(config.Config.Solar.Width), centerY)
	dc.Stroke()
	ref := refTime(sta)
	dc.SetLineWidth(config.Config.Solar.TimeTickLineWidth)
	for i := 0; i < (24 / config.Config.Solar.TimeTickPeriod); i++ {
		tickTime := ref.Add(time.Duration(i*config.Config.Solar.TimeTickPeriod) * time.Hour)
		tickX := timeToPos(sta, tickTime)
		sptick := suncalc.GetPosition(tickTime, config.Config.Lat, config.Config.Lon)
		tickY := altToPos(minAlt, maxAlt, sptick.Altitude)

		dc.DrawLine(tickX, centerY, tickX, tickY)
		dc.Stroke()
	}

	prevX := timeToPos(sta, spdates[0])
	prevY := altToPos(minAlt, maxAlt, sps[0].Altitude)
	dc.SetLineWidth(config.Config.Solar.MarkerLineWidth)
	dc.DrawLine(prevX, prevY, prevX, centerY)
	dc.Stroke()

	dc.SetLineCapSquare()
	dc.SetLineWidth(config.Config.Solar.SunPathLineWidth)
	dc.MoveTo(prevX, prevY)
	prevColor := sunStateColor(spdates[0], stl)
	dc.SetColor(prevColor)
	for i := 1; i < len(sps)+1; i++ {
		nextX := timeToPos(sta, spdates[i%len(sps)])
		nextY := altToPos(minAlt, maxAlt, sps[i%len(sps)].Altitude)
		nextColor := sunStateColor(spdates[i%len(sps)], stl)
		if nextX < prevX {
			dc.LineTo(nextX+float64(config.Config.Solar.Width), nextY)
			dc.MoveTo(prevX-float64(config.Config.Solar.Width), prevY)
		}
		dc.LineTo(nextX, nextY)
		if nextColor != prevColor {
			dc.Stroke()
			dc.SetColor(nextColor)
			dc.MoveTo(nextX, nextY)
		}
		prevX = nextX
		prevY = nextY
		prevColor = nextColor
	}
	dc.Stroke()
	return dc
}

type TimePair struct {
	Time time.Time
	Name suncalc.DayTimeName
}

func main() {
	addHours := 0.
	outPath := ""
	noOut := false
	ignoreCfg := false
	runAsSrv := false
	address := ":8778"
	mode := "lunar"
	flag.Float64Var(&addHours, "forecast", 0, "add hours to time")
	flag.StringVar(&outPath, "out", "out.png", "output path")
	flag.BoolVar(&noOut, "noout", false, "do not output image")
	flag.BoolVar(&ignoreCfg, "ignorecfg", false, "ignorecfg")
	flag.Float64Var(&config.Config.Lat, "lat", config.DefaultConfig.Lat, "user latitude")
	flag.Float64Var(&config.Config.Lon, "lon", config.DefaultConfig.Lon, "user longitude")
	flag.StringVar(&mode, "mode", "lunar", "mode (lunar, solar)")
	flag.BoolVar(&runAsSrv, "server", false, "run as server")
	flag.StringVar(&address, "hostname", ":8778", "server address")
	flag.Parse()

	/*you can do stuff like

	seq 0 1000 | xargs -P 8 -n 1 bash -c '
	  printf -v out "%04d.png" "$0"
	  ./lunar --forecast "$0" -out "$out"
	'
	*/

	config.Init()
	if runAsSrv {
		app := fiber.New()
		app.Get("/lunar/*", func(c fiber.Ctx) error {
			currTime := time.Now().Add(time.Duration(float64(time.Hour) * addHours))
			mp := suncalc.GetMoonPosition(currTime, config.Config.Lat, config.Config.Lon)
			mi := suncalc.GetMoonIllumination(currTime)
			dc := genLunarImage(mp, mi)
			png.Encode(c.Response().BodyWriter(), dc.Image())
			c.Type(".png")
			return c.SendStatus(200)
		})
		app.Get("/solar/*", func(c fiber.Ctx) error {
			currTime := time.Now().Add(time.Duration(float64(time.Hour) * addHours))
			dc := genSolarImage(currTime)
			png.Encode(c.Response().BodyWriter(), dc.Image())
			c.Type(".png")
			return c.SendStatus(200)
		})
		log.Fatal(app.Listen(address))
	} else {
		currTime := time.Now().Add(time.Duration(float64(time.Hour) * addHours))
		if mode == "lunar" {
			mp := suncalc.GetMoonPosition(currTime, config.Config.Lat, config.Config.Lon)
			mi := suncalc.GetMoonIllumination(currTime)

			if !noOut {
				dc := genLunarImage(mp, mi)
				dc.SavePNG(outPath)
			}
			fmt.Printf("∠ %0.2f°, θ %0.2f°", 180./math.Pi*mp.Altitude, 180+180./math.Pi*mp.Azimuth)
		} else {
			sp := suncalc.GetPosition(currTime, config.Config.Lat, config.Config.Lon)

			if !noOut {
				dc := genSolarImage(currTime)
				dc.SavePNG(outPath)
			}
			fmt.Printf("∠ %0.2f°, θ %0.2f°", 180./math.Pi*sp.Altitude, 180+180./math.Pi*sp.Azimuth)

		}
	}
}
