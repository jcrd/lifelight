package main

import (
    "image/color"
    "math/rand"
    "time"

    "github.com/jcrd/go-rpi-rgb-led-matrix"
)

const (
    width = 32
    height = 32
    hardwareMapping = "adafruit-hat"

    tick = time.Second / 6
)

const (
    CELL_DEAD int = iota
    CELL_LIVE
)

type Universe struct {
    cells [width * height]int
    buffer [width * height]int
    matrix rgbmatrix.Matrix
    canvas *rgbmatrix.Canvas
    ticker *time.Ticker
    color color.Color
}

func applyRules(c, n int) int {
    if n == 3 || (n == 2 && c == CELL_LIVE) {
        return CELL_LIVE
    } else {
        return CELL_DEAD
    }
}

func (u *Universe) getIdx(x, y int) int {
    return y * width + x
}

func (u *Universe) getNeighbors(x, y int) int {
    var count int

    for _, w := range []int{width - 1, 0, 1} {
        for _, h := range []int{height - 1, 0, 1} {
            if w == 0 && h == 0 {
                continue
            }
            i := u.getIdx((x + w) % width, (y + h) % height)
            count += u.cells[i]
        }
    }

    return count
}

func (u *Universe) tick() {
    defer u.canvas.Render()

    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            var color color.Color = color.Black
            i := u.getIdx(x, y)
            c := applyRules(u.cells[i], u.getNeighbors(x, y))
            if c == CELL_LIVE {
                color = u.color
            }
            u.buffer[i] = c
            u.canvas.Set(x, y, color)
        }
    }

    u.cells = u.buffer
}

func (u *Universe) randomize() {
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            u.cells[u.getIdx(x, y)] = rand.Intn(2)
        }
    }
}

func (u *Universe) run() {
    for range u.ticker.C {
        u.tick()
    }
}

func (u *Universe) close() {
    u.ticker.Stop()
    u.canvas.Close()
}

func main() {
    config := rgbmatrix.DefaultConfig
    config.Cols = width
    config.Rows = height
    config.HardwareMapping = hardwareMapping

    matrix, err := rgbmatrix.NewRGBLedMatrix(&config)
    if err != nil {
        panic(err)
    }

    rand.Seed(time.Now().UnixNano())

    u := &Universe{
        matrix: matrix,
        canvas: rgbmatrix.NewCanvas(matrix),
        ticker: time.NewTicker(tick),
        color: color.White,
    }
    defer u.close()

    u.randomize()
    u.run()
}
