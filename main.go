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

    ticksPerSecond = 10

    // Set to -1 to disable seeding of dead zones
    seedFrequency = ticksPerSecond * 3

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
    gridSize = width * height
    liveCellN = CELL_N - 1
)

var colorScheme = [CELL_N]color.Color{
    CELL_DEAD: color.Black,
}

type Cells [gridSize]int

type Env struct {
    cells Cells
    buffer Cells
    ticker *time.Ticker
    seedTick int
    deadZones []int
}

func newEnv() *Env {
    return &Env{
        ticker: time.NewTicker(time.Second / ticksPerSecond),
        seedTick: seedFrequency,
        deadZones: make([]int, 0, gridSize),
    }
}

func newCanvas() *rgbmatrix.Canvas {
    config := rgbmatrix.DefaultConfig
    config.Cols = width
    config.Rows = height
    config.HardwareMapping = hardwareMapping

    matrix, err := rgbmatrix.NewRGBLedMatrix(&config)
    if err != nil {
        panic(err)
    }

    return rgbmatrix.NewCanvas(matrix)
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

func randomCell() int {
    if rand.Intn(2) == 1 {
        return rand.Intn(liveCellN) + 1
    }
    return CELL_DEAD
}

func getIdx(x, y int) int {
    return y * width + x
}

func getCoords(idx int) (int, int) {
    return idx % width, idx / width
}

func getNeighbors(idx int) (ns [8]int) {
    x, y := getCoords(idx)
    i := 0

    for _, w := range []int{width - 1, 0, 1} {
        for _, h := range []int{height - 1, 0, 1} {
            if w == 0 && h == 0 {
                continue
            }
            ns[i] = getIdx((x + w) % width, (y + h) % height)
            i++
        }
    }

    return ns
}

func getContext(cells Cells, ns [8]int) (n int, cs [liveCellN]int) {
    for _, i := range ns {
        if c := cells[i]; c != CELL_DEAD {
            n += 1
            cs[c - 1] += 1
        }
    }

    return n, cs
}

func (e *Env) seedDeadZones() {
    e.deadZones = e.deadZones[:0]

    for i := range e.buffer {
        if e.buffer[i] != CELL_DEAD {
            continue
        }
        ns := getNeighbors(i)
        if n, _ := getContext(e.buffer, ns); n == 0 {
            e.deadZones = append(e.deadZones, i)
        }
    }

    i := e.deadZones[rand.Intn(len(e.deadZones))]
    e.buffer[i] = randomCell()

    for _, n := range getNeighbors(i) {
        e.buffer[n] = randomCell()
    }
}

func (e *Env) tick() Cells {
    for i := range e.buffer {
        n, cs := getContext(e.cells, getNeighbors(i))
        e.buffer[i] = applyRules(e.cells[i], n, cs)
    }

    if e.seedTick > 0 {
        e.seedTick--
    }
    if e.seedTick == 0 {
        e.seedDeadZones()
        e.seedTick = seedFrequency
    }

    e.cells = e.buffer
    return e.cells
}

func (e *Env) randomize() {
    for i := range e.cells {
        e.cells[i] = randomCell()
    }
}

func (e *Env) run(canvas *rgbmatrix.Canvas) {
    for range e.ticker.C {
        for i, c := range e.tick() {
            x, y := getCoords(i)
            canvas.Set(x, y, colorScheme[c])
        }
        canvas.Render()
    }
}

func (e *Env) close() {
    e.ticker.Stop()
}

func main() {
    canvas := newCanvas()
    defer canvas.Close()

    e := newEnv()
    defer e.close()

    rand.Seed(time.Now().UnixNano())
    genColors(fastColorGen)

    e.randomize()
    e.run(canvas)
}
