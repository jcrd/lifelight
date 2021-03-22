package main

import (
    "image/color"
    "log"
    "math/rand"
    "os"
    "strings"
    "time"

    "github.com/jcrd/go-rpi-rgb-led-matrix"
    "github.com/lucasb-eyer/go-colorful"
)

const (
    CELL_DEAD int = iota
    CELL_LIVE_1
    CELL_LIVE_2
    CELL_LIVE_3
    CELL_LIVE_4

    CELL_N
)

const liveCellN = CELL_N - 1

var colorScheme = [CELL_N]color.Color{
    CELL_DEAD: color.Black,
}
var debug bool

type Cells []int
type Neighbors [8]int

type Env struct {
    cells Cells
    buffer Cells
    deadZones Cells
    width int
    height int
    size int
    seedThreshold float32
    seedThresholdDecayTicks int
    seedCooldownTicks int
    config *Config
}

func debugLog(fmt string, v ...interface{}) {
    if debug {
        log.Printf(fmt, v...)
    }
}

func getDebug() bool {
    _, ok := os.LookupEnv("LIFELIGHT_DEBUG")
    if ok {
        log.Println("Debug logging enabled")
    }
    return ok
}

func newEnv(c *Config) *Env {
    size := c.Hardware.MatrixWidth * c.Hardware.MatrixHeight

    return &Env{
        cells: make(Cells, size),
        buffer: make(Cells, size),
        deadZones: make(Cells, 0, size),
        width: c.Hardware.MatrixWidth,
        height: c.Hardware.MatrixHeight,
        size: size,
        seedThreshold: c.SeedThreshold,
        seedThresholdDecayTicks: 0,
        seedCooldownTicks: 0,
        config: c,
    }
}

func newCanvas(c *Config) *rgbmatrix.Canvas {
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

    debugLog("Generating color scheme...\n")

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

func getIdx(x, y, width int) int {
    return y * width + x
}

func getCoords(idx, width int) (int, int) {
    return idx % width, idx / width
}

func getNeighbors(idx, width, height int) (ns Neighbors) {
    x, y := getCoords(idx, width)
    i := 0

    for _, w := range [...]int{width - 1, 0, 1} {
        for _, h := range [...]int{height - 1, 0, 1} {
            if w == 0 && h == 0 {
                continue
            }
            ns[i] = getIdx((x + w) % width, (y + h) % height, width)
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

func (e *Env) getNeighbors(idx int) Neighbors {
    return getNeighbors(idx, e.width, e.height)
}

func (e *Env) updateDeadZones() int {
    e.deadZones = e.deadZones[:0]

    for i := range e.buffer {
        if e.buffer[i] != CELL_DEAD {
            continue
        }
        ns := e.getNeighbors(i)
        if n, _ := getContext(e.buffer, ns); n == 0 {
            e.deadZones = append(e.deadZones, i)
        }
    }

    return len(e.deadZones)
}

func (e *Env) seedDeadZones() {
    i := e.deadZones[rand.Intn(len(e.deadZones))]
    e.buffer[i] = randomCell()

    for _, n := range e.getNeighbors(i) {
        e.buffer[n] = randomCell()
    }
}

func (e *Env) seed() {
    if e.seedCooldownTicks > 0 {
        e.seedCooldownTicks--
        debugLog("seed: cooldown = %d\n", e.seedCooldownTicks)
        return
    }

    z := e.updateDeadZones()
    if z == 0 {
        return
    }

    t := float32(z) / float32(e.size)
    c := e.config

    if t >= e.seedThreshold || e.seedThreshold < c.SeedThresholdDecay {
        debugLog("seed: deadzones = %f; seeding...\n", t)
        e.seedDeadZones()
        e.seedThreshold = c.SeedThreshold
        e.seedThresholdDecayTicks = c.SeedThresholdDecayTicks
        e.seedCooldownTicks = c.SeedCooldownTicks
    } else if e.seedThresholdDecayTicks > 0 {
        e.seedThresholdDecayTicks--
        debugLog("seed: decay = %d\n", e.seedThresholdDecayTicks)
    } else {
        e.seedThreshold -= e.config.SeedThresholdDecay
        e.seedThresholdDecayTicks = c.SeedThresholdDecayTicks
        debugLog("seed: threshold = %f\n", e.seedThreshold)
    }
}

func (e *Env) tick() Cells {
    for i := range e.buffer {
        n, cs := getContext(e.cells, e.getNeighbors(i))
        e.buffer[i] = applyRules(e.cells[i], n, cs)
    }
    e.seed()
    copy(e.cells, e.buffer)

    return e.cells
}

func (e *Env) randomize() {
    for i := range e.cells {
        e.cells[i] = randomCell()
    }
}

func (e *Env) update(canvas *rgbmatrix.Canvas) {
    for i, c := range e.tick() {
        x, y := getCoords(i, e.width)
        canvas.Set(x, y, colorScheme[c])
    }
    canvas.Render()
}

func (e *Env) clear(canvas *rgbmatrix.Canvas) {
    for x := 0; x < e.width; x++ {
        for y := 0; y < e.height; y++ {
            canvas.Set(x, y, color.Black)
        }
    }
    canvas.Render()
}

func main() {
    debug = getDebug()

    c := newConfig()
    c.load()

    canvas := newCanvas(c)
    defer canvas.Close()

    rand.Seed(time.Now().UnixNano())
    genColors(c.FastColorGen)

    e := newEnv(c)
    e.randomize()

    ticks := time.Second / time.Duration(c.TicksPerSecond)
    state := true
    toggle := make(chan struct{})

    go func() {
        for range time.Tick(time.Second * 30) {
            t := strings.Fields(time.Now().Format("Mon 15:04"))
            if s := c.getScheduleState(t[0], t[1], state); s != state {
                state = s
                toggle <- struct{}{}
            }
        }
    }()

    for range time.Tick(ticks) {
        select {
        case <-toggle:
            e.clear(canvas)
            <-toggle
        default:
            e.update(canvas)
        }
    }
}
