package main

import (
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

var configPath = "/etc/lifelight.ini"

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

func genColors(fast bool) {
    var colors []colorful.Color

regen:
    if !fast {
        var err error
        colors, err = colorful.HappyPalette(life.LiveCellN)
        if err != nil {
            fast = true
            goto regen
        }
    } else {
        colors = colorful.FastHappyPalette(life.LiveCellN)
    }

    cs := life.ColorScheme{
        color.Black,
    }

    for i, c := range colors {
        cs[i + 1] = c
    }

    life.SetColorScheme(cs)
}

func main() {
    life.InitLogger(os.Getenv("LIFELIGHT_DEBUG"))

    if v, hasEnv := os.LookupEnv("LIFELIGHT_CONFIG"); hasEnv {
        if _, err := os.Stat(v); err != nil {
            log.Printf("config: %v\n", err)
            return
        }
        configPath = v
    }

    c := life.NewConfig()
    if err := c.Load(configPath); err != nil {
        log.Printf("config: %v\n", err)
        return
    }

    canvas := newCanvas(c)
    defer canvas.Close()

    rand.Seed(time.Now().UnixNano())
    genColors(c.FastColorGen)

    e := life.NewEnv(c)
    e.Randomize()

    ticks := time.Second / time.Duration(c.TicksPerSecond)
    ticker := time.NewTicker(ticks)
    defer ticker.Stop()

    state := true
    toggle := make(chan struct{})

    updateState := func() {
        t := strings.Fields(time.Now().Format("Mon 15:04"))
        if s := c.GetScheduleState(t[0], t[1], state); s != state {
            state = s
            toggle <- struct{}{}
        }
    }

    go func() {
        updateState()
        for range time.Tick(time.Second * 30) {
            updateState()
        }
    }()

    for {
        select {
        case <-toggle:
            e.Clear(canvas)
            <-toggle
        case <-ticker.C:
            e.Update(canvas)
        }
    }
}
