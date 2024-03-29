package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"lifelight/life"

	"github.com/jcrd/go-rpi-rgb-led-matrix"
	"github.com/lucasb-eyer/go-colorful"
)

var version = ""

var configPath = "/etc/lifelight.ini"

var colorPalettes = map[string]func(int) ([]colorful.Color, error){
	"happy": colorful.HappyPalette,
	"soft":  colorful.SoftPalette,
	"warm":  colorful.WarmPalette,
}

func newCanvas(c *life.Config) *rgbmatrix.Canvas {
	config := rgbmatrix.DefaultConfig
	config.Cols = c.Hardware.MatrixWidth
	config.Rows = c.Hardware.MatrixHeight
	config.HardwareMapping = c.Hardware.Mapping

	matrix, err := rgbmatrix.NewRGBLedMatrix(&config)
	if err != nil {
		panic(err)
	}

	return rgbmatrix.NewCanvas(matrix)
}

func makeColorScheme(hs []string) (life.ColorScheme, error) {
	cs := life.ColorScheme{
		color.Black,
	}

	for i, h := range hs {
		c, err := colorful.Hex(h)
		if err != nil {
			return cs, fmt.Errorf("%s (%v)", h, err)
		}
		cs[i+1] = c
	}

	return cs, nil
}

func genColors(ps []string) {
	p := ps[rand.Intn(len(ps))]
	colors, err := colorPalettes[p](life.LiveCellN)

	if err != nil {
		log.Printf("color: Failed to generate palette '%s': %v\n", p, err)
		colors = colorful.FastHappyPalette(life.LiveCellN)
	}

	cs := life.ColorScheme{
		color.Black,
	}

	for i, c := range colors {
		cs[i+1] = c
	}

	life.SetColorScheme(cs)
}

func initialState(c *life.Config) bool {
	t := time.Now()

	for i := 1; i < 8; i++ {
		p := t.AddDate(0, 0, -i)
		d := p.Format("Mon")
		if c.HasSchedule(d) {
			return c.GetScheduleState(d, "23:59", true)
		}
	}

	return true
}

func updateScheduleState(c *life.Config, toggle chan<- struct{}) {
	running := true
	state := initialState(c)

	update := func() {
		t := strings.Fields(time.Now().Format("Mon 15:04"))
		if state = c.GetScheduleState(t[0], t[1], state); running != state {
			running = state
			toggle <- struct{}{}
		}
	}

	update()

	for range time.Tick(time.Second * 30) {
		update()
	}
}

func main() {
	if v := os.Getenv("LIFELIGHT_DEBUG"); v != "" {
		life.InitLogger(v)
	}

	v, hasEnv := os.LookupEnv("LIFELIGHT_CONFIG")
	if hasEnv {
		configPath = v
	}

	c := life.NewConfig()

	if err := c.Load(configPath, hasEnv); err != nil {
		log.Printf("config: %v\n", err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	if len(c.Color.Scheme) > 0 {
		cs, err := makeColorScheme(c.Color.Scheme)
		if err != nil {
			log.Printf("config: Failed to parse color: %v\n", err)
			return
		}
		life.SetColorScheme(cs)
		c.Color.ScheduleRegen = false
	} else {
		genColors(c.Color.Palettes)
	}

	canvas := newCanvas(c)
	defer canvas.Close()

	e := life.NewEnv(c)
	e.Randomize()

	ticks := time.Second / time.Duration(c.TicksPerSecond)
	ticker := time.NewTicker(ticks)
	defer ticker.Stop()

	toggle := make(chan struct{})

	fmt.Println("running:", version)

	if c.Schedule && c.NumSchedules() > 0 {
		go updateScheduleState(c, toggle)
	}

	for {
		select {
		case <-toggle:
			e.Clear(canvas)
			if c.Color.ScheduleRegen {
				genColors(c.Color.Palettes)
			}
			<-toggle
		case <-ticker.C:
			e.Update(canvas)
		}
	}
}
