package life

import (
    "image/color"
    "math/rand"
)

const (
    CELL_DEAD int = iota
    CELL_LIVE_1
    CELL_LIVE_2
    CELL_LIVE_3
    CELL_LIVE_4

    CELL_N
)

const LiveCellN = CELL_N - 1

var colorScheme = ColorScheme{
    color.Black,
    color.RGBA{255, 0, 0, 255},
    color.RGBA{0, 255, 0, 255},
    color.RGBA{0, 0, 255, 255},
    color.White,
}

type Logger interface {
    init(string)
    log(string, string, ...interface{})
}

var logger Logger

type ColorScheme [CELL_N]color.Color
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

type Renderer interface {
    Set(int, int, color.Color)
    Render() error
}

func InitLogger(domains string) {
    logger.init(domains)
}

func SetColorScheme(cs ColorScheme) {
    colorScheme = cs
}

func NewEnv(c *Config) *Env {
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

func applyRules(c, n int, cs [LiveCellN]int) int {
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
        return rand.Intn(LiveCellN) + 1
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

func getContext(cells Cells, ns [8]int) (n int, cs [LiveCellN]int) {
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
        logger.log("seed", "cooldown = %d\n", e.seedCooldownTicks)
        return
    }

    z := e.updateDeadZones()
    if z == 0 {
        return
    }

    t := float32(z) / float32(e.size)
    c := e.config

    if t >= e.seedThreshold || e.seedThreshold < c.SeedThresholdDecay {
        logger.log("seed", "deadzones = %f; seeding...\n", t)
        e.seedDeadZones()
        e.seedThreshold = c.SeedThreshold
        e.seedThresholdDecayTicks = c.SeedThresholdDecayTicks
        e.seedCooldownTicks = c.SeedCooldownTicks
    } else if e.seedThresholdDecayTicks > 0 {
        e.seedThresholdDecayTicks--
        logger.log("seed", "decay = %d\n", e.seedThresholdDecayTicks)
    } else {
        e.seedThreshold -= e.config.SeedThresholdDecay
        e.seedThresholdDecayTicks = c.SeedThresholdDecayTicks
        logger.log("seed", "threshold = %f\n", e.seedThreshold)
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

func (e *Env) Randomize() {
    for i := range e.cells {
        e.cells[i] = randomCell()
    }
}

func (e *Env) Update(r Renderer) {
    for i, c := range e.tick() {
        x, y := getCoords(i, e.width)
        r.Set(x, y, colorScheme[c])
    }
    r.Render()
}

func (e *Env) Clear(r Renderer) {
    for x := 0; x < e.width; x++ {
        for y := 0; y < e.height; y++ {
            r.Set(x, y, color.Black)
        }
    }
    r.Render()
}
