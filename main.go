package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"math"
	"time"

	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v3"
	"github.com/sixdouglas/suncalc"

	"lunar/config"
)

const moonRadius = 1737.4 // km

func screenCoords(x, y, visAngle float64) (float64, float64) {
	return float64(config.Config.ImageSize)/2 + config.Config.MoonRad*(math.Cos(visAngle)*x-math.Sin(visAngle)*y), float64(config.Config.ImageSize)/2 + config.Config.MoonRad*(math.Sin(visAngle)*x+math.Cos(visAngle)*y)
}

func SignedAngleDiff(a, b float64) float64 {
	diff := math.Mod(b-a+math.Pi, 2*math.Pi)
	if diff < 0 {
		diff += 2 * math.Pi
	}
	return diff - math.Pi
}

func genImage(mp suncalc.MoonPosition, mi suncalc.MoonIllumination) *gg.Context {
	visAngle := SignedAngleDiff(mi.Angle, mp.ParallacticAngle)
	//fmt.Println(mp.ParallacticAngle, mi.Angle)
	anglePhase := mi.Phase * (2 * math.Pi)
	//fmt.Println(visAngle*360/(2*math.Pi), mi.Phase)
	dc := gg.NewContext(config.Config.ImageSize, config.Config.ImageSize)
	dc.SetLineWidth(config.Config.LineWidth)
	dc.DrawCircle(float64(config.Config.ImageSize)/2, float64(config.Config.ImageSize)/2, config.Config.MoonRad)
	dc.SetColor(config.OutColor)
	dc.Stroke()
	if config.Config.Parallels || config.Config.Meridians {
		maskdc := gg.NewContext(config.Config.ImageSize, config.Config.ImageSize)
		maskdc.SetLineWidth(config.Config.GeodesicsThickness)

		if config.Config.Parallels {
			for i := 0; i <= config.Config.ParallelCount; i++ {
				phi := math.Pi * float64(i+1) / float64(config.Config.ParallelCount+2)
				currX, currY := screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(0), visAngle)
				maskdc.MoveTo(currX, currY)
				for i := 0; i < config.Config.LineRes; i++ {
					theta := math.Pi * float64(i+1) / float64(config.Config.LineRes)
					currX, currY = screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(theta), visAngle)
					maskdc.LineTo(currX, currY)
				}
				maskdc.Stroke()
			}
		}
		if config.Config.Meridians {
			for i := 0; i <= config.Config.MeridianCount; i++ {
				theta := math.Pi * float64(i+1) / float64(config.Config.MeridianCount+2)
				currX, currY := screenCoords(math.Cos(0), math.Sin(0)*math.Cos(theta), visAngle)
				maskdc.MoveTo(currX, currY)
				for i := 0; i < config.Config.LineRes; i++ {
					phi := math.Pi * float64(i+1) / float64(config.Config.LineRes)
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
	for i := 0; i < config.Config.LineRes; i++ {
		phi := math.Pi * float64(i+1) / float64(config.Config.LineRes)
		if math.Sin(anglePhase) > 0 {
			currX, currY = screenCoords(math.Cos(phi), math.Sin(phi)*math.Cos(anglePhase), visAngle) // Should be opposite if sin < 0 for anglePhase
		} else {
			currX, currY = screenCoords(math.Cos(phi), -math.Sin(phi)*math.Cos(anglePhase), visAngle) // Should be opposite if sin < 0 for anglePhase
		}
		dc.LineTo(currX, currY)
	}
	for i := 0; i < config.Config.LineRes; i++ {
		phi := math.Pi * (1 - float64(i+1)/float64(config.Config.LineRes))
		if math.Sin(anglePhase) > 0 {
			currX, currY = screenCoords(math.Cos(phi), math.Sin(phi), visAngle) // Should be opposite if sin < 0 for anglePhase
		} else {
			currX, currY = screenCoords(math.Cos(phi), -math.Sin(phi), visAngle) // Should be opposite if sin < 0 for anglePhase
		}
		dc.LineTo(currX, currY)
	}
	dc.Fill()

	dc.SetLineWidth(config.Config.HorizonLineWidth)
	dc.SetDash(5, 10)
	frac := mp.Altitude / (moonRadius / mp.Distance / config.Config.MoonRad * float64(config.Config.ImageSize))
	if (frac <= 1) && mp.Altitude >= 0 {
		partFrac := (frac*float64(config.Config.ImageSize) - (float64(config.Config.ImageSize)/2 - config.Config.MoonRad)) / (2 * config.Config.MoonRad)
		if partFrac >= 0 && partFrac <= 1 {
			// two lines
			y := math.Sqrt(1. - (partFrac-1./2.)*(partFrac-1./2.)*4.)
			dc.DrawLine(config.Config.HorizonLinePadding, float64(config.Config.ImageSize)*(1-frac), float64(config.Config.ImageSize/2)-y*config.Config.MoonRad, float64(config.Config.ImageSize)*(1-frac))
			dc.DrawLine(float64(config.Config.ImageSize)-config.Config.HorizonLinePadding, float64(config.Config.ImageSize)*(1-frac), float64(config.Config.ImageSize/2)+y*config.Config.MoonRad, float64(config.Config.ImageSize)*(1-frac))
		} else {
			// single line
			dc.DrawLine(config.Config.HorizonLinePadding, float64(config.Config.ImageSize)*(1-frac), float64(config.Config.ImageSize)-config.Config.HorizonLinePadding, float64(config.Config.ImageSize)*(1-frac))
		}
	}
	dc.Stroke()
	return dc
}

func main() {
	addHours := 0.
	outPath := ""
	noOut := false
	ignoreCfg := false
	runAsSrv := false
	address := ":8778"
	flag.Float64Var(&addHours, "forecast", 0, "add hours to time")
	flag.StringVar(&outPath, "out", "out.png", "output path")
	flag.BoolVar(&noOut, "noout", false, "do not output image")
	flag.BoolVar(&ignoreCfg, "ignorecfg", false, "ignorecfg")
	flag.Float64Var(&config.Config.Lat, "lat", config.DefaultConfig.Lat, "user latitude")
	flag.Float64Var(&config.Config.Lon, "lon", config.DefaultConfig.Lon, "user longitude")
	flag.IntVar(&config.Config.ImageSize, "imagesize", config.DefaultConfig.ImageSize, "image size")
	flag.Float64Var(&config.Config.MoonRad, "moonrad", config.DefaultConfig.MoonRad, "moon radius in pixels")
	flag.Float64Var(&config.Config.LineWidth, "linewidth", config.DefaultConfig.LineWidth, "line width in pixels")
	flag.IntVar(&config.Config.LineRes, "lineres", config.DefaultConfig.LineRes, "line res in steps")
	flag.StringVar(&config.Config.Color, "color", config.DefaultConfig.Color, "draw color")
	flag.BoolVar(&config.Config.HorizonLine, "horizonline", config.DefaultConfig.HorizonLine, "draw horizon line")
	flag.Float64Var(&config.Config.HorizonLineWidth, "horizonlinewidth", config.DefaultConfig.HorizonLineWidth, "horizon line width")
	flag.Float64Var(&config.Config.HorizonLinePadding, "horizonlinepadding", config.DefaultConfig.HorizonLinePadding, "horizon line padding")
	flag.BoolVar(&config.Config.Meridians, "meridians", config.DefaultConfig.Meridians, "draw meridians")
	flag.BoolVar(&config.Config.Parallels, "parallels", config.DefaultConfig.Parallels, "draw parallels")
	flag.Float64Var(&config.Config.GeodesicsThickness, "geodesicsthickness", config.DefaultConfig.GeodesicsThickness, "geodesics thickness")
	flag.IntVar(&config.Config.MeridianCount, "meridiancount", config.DefaultConfig.MeridianCount, "meridian count")
	flag.IntVar(&config.Config.ParallelCount, "parallelcount", config.DefaultConfig.ParallelCount, "parallel count")
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
		app.Get("/*", func(c fiber.Ctx) error {
			currTime := time.Now().Add(time.Duration(float64(time.Hour) * addHours))
			mp := suncalc.GetMoonPosition(currTime, config.Config.Lat, config.Config.Lon)
			mi := suncalc.GetMoonIllumination(currTime)
			dc := genImage(mp, mi)
			png.Encode(c.Response().BodyWriter(), dc.Image())
			c.Type(".png")
			return c.SendStatus(200)
		})
		log.Fatal(app.Listen(address))
	} else {
		currTime := time.Now().Add(time.Duration(float64(time.Hour) * addHours))
		mp := suncalc.GetMoonPosition(currTime, config.Config.Lat, config.Config.Lon)
		mi := suncalc.GetMoonIllumination(currTime)

		if !noOut {
			dc := genImage(mp, mi)
			dc.SavePNG(outPath)
		}
		fmt.Printf("∠ %0.2f°, θ %0.2f°", 180./math.Pi*mp.Altitude, 180+180./math.Pi*mp.Azimuth)
	}
}
