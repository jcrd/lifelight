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

type Universe struct {
    cells Cells
    buffer Cells
    matrix rgbmatrix.Matrix
    canvas *rgbmatrix.Canvas
    ticker *time.Ticker
    seedTick int
    deadZones []int
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

func (u *Universe) seedDeadZones() {
    u.deadZones = u.deadZones[:0]

    for i := range u.buffer {
        if u.buffer[i] != CELL_DEAD {
            continue
        }
        ns := getNeighbors(i)
        if n, _ := getContext(u.buffer, ns); n == 0 {
            u.deadZones = append(u.deadZones, i)
        }
    }

    i := u.deadZones[rand.Intn(len(u.deadZones))]
    u.buffer[i] = randomCell()

    for _, n := range getNeighbors(i) {
        u.buffer[n] = randomCell()
    }
}

func (u *Universe) tick() {
    for i := range u.buffer {
        n, cs := getContext(u.cells, getNeighbors(i))
        u.buffer[i] = applyRules(u.cells[i], n, cs)
    }

    if u.seedTick > 0 {
        u.seedTick--
    }
    if u.seedTick == 0 {
        u.seedDeadZones()
        u.seedTick = seedFrequency
    }

    for i, c := range u.buffer {
        x, y := getCoords(i)
        u.canvas.Set(x, y, colorScheme[c])
    }

    u.cells = u.buffer
    u.canvas.Render()
}

func (u *Universe) randomize() {
    for i := range u.cells {
        u.cells[i] = randomCell()
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
        ticker: time.NewTicker(time.Second / ticksPerSecond),
        seedTick: seedFrequency,
        deadZones: make([]int, 0, gridSize),
    }
    defer u.close()

    genColors(fastColorGen)

    u.randomize()
    u.run()
}
