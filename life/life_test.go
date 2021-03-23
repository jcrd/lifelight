package life

import (
    "testing"
)

const (
    testWidth = 32
    testHeight = 32
)

var (
    glider0 = Cells{
        0, 0, 0, 0, 0,
        0, 0, 1, 0, 0,
        0, 0, 0, 2, 0,
        0, 1, 3, 3, 0,
        0, 0, 0, 0, 0,
    }
    glider1 = Cells{
        0, 0, 0, 0, 0,
        0, 0, 0, 0, 0,
        0, 1, 0, 2, 0,
        0, 0, 3, 3, 0,
        0, 0, 3, 0, 0,
    }
    gliderWidth = 5
    gliderHeight = 5
)

func testGetNeighbors(t *testing.T, idx int, vals []int) {
    ns := getNeighbors(idx, testWidth, testHeight)
    for i, v := range vals {
        if ns[i] != v {
            t.Errorf("neighbors[%d] = %d; want %d", i, ns[i], v)
        }
    }
}

func TestGetIdx(t *testing.T) {
    i := getIdx(24, 12, testWidth)
    if i != 408 {
        t.Errorf("getIdx = %d; want 408", i)
    }
}

func TestGetCoords(t *testing.T) {
    x, y := getCoords(408, testWidth)
    if x != 24 || y != 12 {
        t.Errorf("x = %d, y = %d; want 24, 12", x, y)
    }
}

func TestGetNeighbors(t *testing.T) {
    testGetNeighbors(t, 408, []int{375, 407, 439, 376, 440, 377, 409, 441})
}

func TestGetNeighborsEdge(t *testing.T) {
    testGetNeighbors(t, 1023, []int{990, 1022, 30, 991, 31, 960, 992, 0})
}

func TestGetContext(t *testing.T) {
    idx := getIdx(2, 2, gliderWidth)
    ns := getNeighbors(idx, gliderWidth, gliderHeight)
    n, cs := getContext(glider0, ns)

    if n != 5 {
        t.Errorf("n = %d; want 5", n)
    }

    for i, v := range [...]int{2, 1, 2, 0} {
        if cs[i] != v {
            t.Errorf("count = %d; want %d", cs[i], v)
        }
    }
}

func TestApplyRules(t *testing.T) {
    cells := make(Cells, gliderWidth * gliderHeight)

    for i, c := range glider0 {
        ns := getNeighbors(i, gliderWidth, gliderHeight)
        n, cs := getContext(glider0, ns)
        cells[i] = applyRules(c, n, cs)
    }

    for i, c := range cells {
        if c != glider1[i] {
            t.Errorf("cell = %d; want %d", c, glider1[i])
        }
    }
}
