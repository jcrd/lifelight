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
    CELL_LIVE_1
    CELL_LIVE_2
    CELL_LIVE_3
    CELL_LIVE_4

    CELL_N
)

var colorScheme = [CELL_N]color.Color{
    CELL_DEAD: color.Black,
    CELL_LIVE_1: color.RGBA{255, 0, 0, 255},
    CELL_LIVE_2: color.RGBA{0, 255, 0, 255},
    CELL_LIVE_3: color.RGBA{0, 0, 255, 255},
    CELL_LIVE_4: color.White,
}

type Universe struct {
    cells [width * height]int
    buffer [width * height]int
    matrix rgbmatrix.Matrix
    canvas *rgbmatrix.Canvas
    ticker *time.Ticker
}

func applyRules(c, n int, cs [CELL_N]int) int {
    if n > 3 || n < 2 {
        return CELL_DEAD
    }

    if c == CELL_DEAD && n == 3 {
        for s, i := range cs[1:] {
            s += 1
            if i > 1 {
                return s
            }
            if i == 0 {
                c = s
            }
        }
    }

    return c
}

func (u *Universe) getIdx(x, y int) int {
    return y * width + x
}

func (u *Universe) getNeighbors(x, y int) (n int, cs [CELL_N]int) {
    for _, w := range []int{width - 1, 0, 1} {
        for _, h := range []int{height - 1, 0, 1} {
            if w == 0 && h == 0 {
                continue
            }
            i := u.getIdx((x + w) % width, (y + h) % height)
            cs[u.cells[i]] += 1
        }
    }

    for _, i := range cs[1:] {
        n += i
    }

    return n, cs
}

func (u *Universe) tick() {
    defer u.canvas.Render()

    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            i := u.getIdx(x, y)
            n, cs := u.getNeighbors(x, y)
            c := applyRules(u.cells[i], n, cs)
            u.buffer[i] = c
            u.canvas.Set(x, y, colorScheme[c])
        }
    }

    u.cells = u.buffer
}

func (u *Universe) randomize() {
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            s := CELL_DEAD
            if rand.Intn(2) == 1 {
                s = rand.Intn(CELL_N - 1) + 1
            }
            u.cells[u.getIdx(x, y)] = s
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
    }
    defer u.close()

    u.randomize()
    u.run()
}
