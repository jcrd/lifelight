package main

import (
    "image/color"
    "log"
    "math/rand"
    "time"

    "github.com/jcrd/go-rpi-rgb-led-matrix"
    "github.com/lucasb-eyer/go-colorful"
)

const (
    width = 32
    height = 32
    hardwareMapping = "adafruit-hat"

    tick = time.Second / 6

    fastColorGen = true
)

const (
    CELL_DEAD int = iota
    CELL_LIVE_1
    CELL_LIVE_2
    CELL_LIVE_3
    CELL_LIVE_4

    CELL_N
)

const (
    liveCellN = CELL_N - 1
)

var colorScheme = [CELL_N]color.Color{
    CELL_DEAD: color.Black,
}

type Universe struct {
    cells [width * height]int
    buffer [width * height]int
    matrix rgbmatrix.Matrix
    canvas *rgbmatrix.Canvas
    ticker *time.Ticker
}

func genColors(fast bool) {
    var colors []colorful.Color

    log.Println("Generating color scheme...")

regen:
    if !fast {
        var err error
        colors, err = colorful.HappyPalette(liveCellN)
        if err != nil {
            fast = true
            goto regen
        }
    } else {
        colors = colorful.FastHappyPalette(liveCellN)
    }

    for i, c := range colors {
        colorScheme[i + 1] = c
    }
}

func applyRules(c, n int, cs [liveCellN]int) int {
    if n > 3 || n < 2 {
        return CELL_DEAD
    }

    if c == CELL_DEAD && n == 3 {
        for s, i := range cs {
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

func (u *Universe) getNeighbors(x, y int) (n int, cs [liveCellN]int) {
    for _, w := range []int{width - 1, 0, 1} {
        for _, h := range []int{height - 1, 0, 1} {
            if w == 0 && h == 0 {
                continue
            }
            i := u.getIdx((x + w) % width, (y + h) % height)
            if c := u.cells[i]; c != CELL_DEAD {
                n += 1
                cs[c - 1] += 1
            }
        }
    }

    return n, cs
}

func (u *Universe) tick() {
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
    u.canvas.Render()
}

func (u *Universe) randomize() {
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            s := CELL_DEAD
            if rand.Intn(2) == 1 {
                s = rand.Intn(liveCellN) + 1
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

    genColors(fastColorGen)

    u.randomize()
    u.run()
}
