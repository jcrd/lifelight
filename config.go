package main

import (
    "errors"
    "fmt"
    "log"
    "os"
    "sort"
    "strings"

    "github.com/go-ini/ini"
)

const configPath = "/etc/lifelight.ini"

var hardwareMappings = []string{
    "regular",
    "adafruit-hat",
    "adafruit-hat-pwm",
    "compute-module",
}

type Time struct {
    state bool
    hh int
    mm int
}

type Hardware struct {
    MatrixWidth int
    MatrixHeight int
    Mapping string
}

type Config struct {
    TicksPerSecond int
    SeedThreshold float32
    SeedThresholdDecay float32
    SeedThresholdDecayTicks int
    SeedCooldownTicks int
    FastColorGen bool

    Hardware

    schedules map[string][]Time
}

func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}

func parseTime(str string) (t Time, err error) {
    _, err = fmt.Sscanf(str, "%2d:%2d", &t.hh, &t.mm)
    return t, err
}

func toTime(str string, state bool) Time {
    t, _ := parseTime(str)
    t.state = state
    return t
}

func (t Time) Compare(a Time) int {
    if t.hh < a.hh {
        return -1
    }
    if t.hh == a.hh && t.mm == a.mm {
        return 0
    }
    if t.hh > a.hh || t.mm >= a.mm {
        return 1
    }
    return -1
}

func (t Time) String() string {
    return fmt.Sprintf("%02d:%02d", t.hh, t.mm)
}

func newConfig() *Config {
    c := &Config{
        TicksPerSecond: 10,
        SeedThreshold: 0.5,
        SeedThresholdDecay: 0.05,
        SeedThresholdDecayTicks: 5,
        SeedCooldownTicks: 2,
        FastColorGen: true,
        Hardware: Hardware{
            MatrixWidth: 32,
            MatrixHeight: 32,
            Mapping: "adafruit-hat",
        },
    }
    c.schedules = make(map[string][]Time)

    return c
}

func (c *Config) loadSchedules(f *ini.File) {
    var times [2]string

sections:
    for _, section := range f.ChildSections("Schedule") {
        days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
        name := section.Name()

        if k, err := section.GetKey("Days"); err == nil {
            ds := k.Strings(",")
            for _, d := range ds {
                if !contains(days, d) {
                    log.Printf("config: section '%s': Invalid day '%s'\n",
                        name, d)
                    continue sections
                }
            }
            days = ds
        }

        for i, n := range [...]string{"Off", "On"} {
            k, err := section.GetKey(n)
            if err != nil {
                log.Printf("config: section '%s': Requires key '%s'\n",
                    name, n)
                continue sections
            }
            times[i] = k.Value()
        }

        for i, str := range times {
            t, err := parseTime(str)
            if err != nil {
                log.Printf("config: section '%s': Invalid time '%s'\n",
                    name, str)
                continue sections
            }
            t.state = i == 1

            for _, d := range days {
                if _, ok := c.schedules[d]; !ok {
                    c.schedules[d] = make([]Time, 0)
                }
                c.schedules[d] = append(c.schedules[d], t)
            }
        }

        for _, ts := range c.schedules {
            sort.Slice(ts, func(i, j int) bool {
                return ts[j].Compare(ts[i]) == 1
            })
        }

        if debug {
            log.Printf("config: %s: on %s, off %s (%s)\n",
                name, times[1], times[0], strings.Join(days, ", "))
        }
    }
}

func (c *Config) load() error {
    path := configPath
    v, hasEnv := os.LookupEnv("LIFELIGHT_CONFIG")
    if hasEnv {
        path = v
    }

    if _, err := os.Stat(path); err != nil {
        if hasEnv {
            return err
        }
        return nil
    }

    debugLog("Loading config file...\n")

    f, err := ini.Load(path)
    if err != nil {
        return err
    }

    if err = f.MapTo(c); err != nil {
        e := fmt.Sprintf("Failed to parse file '%s': %v", path, err)
        return errors.New(e)
    }

    if c.TicksPerSecond < 1 {
        e := fmt.Sprintf("TicksPerSecond = %d; must be > 0",
            c.TicksPerSecond)
        return errors.New(e)
    }
    if c.SeedThreshold > 1.0 || c.SeedThreshold < 0.0 {
        e := fmt.Sprintf("SeedThreshold = %f; must be in range [0.0, 1.0]",
            c.SeedThreshold)
        return errors.New(e)
    }
    if c.SeedThresholdDecay < 0.0 {
        e := fmt.Sprintf("SeedThresholdDecay = %f; must be positive",
            c.SeedThresholdDecay)
        return errors.New(e)
    }
    if c.SeedThresholdDecayTicks < 0 {
        e := fmt.Sprintf("SeedThresholdDecayTicks = %d; must be positive",
            c.SeedThresholdDecayTicks)
        return errors.New(e)
    }
    if c.SeedCooldownTicks < 0 {
        e := fmt.Sprintf("SeedCooldownTicks = %d; must be positive",
            c.SeedCooldownTicks)
        return errors.New(e)
    }
    if c.Hardware.MatrixWidth < 1 {
        e := fmt.Sprintf("Hardware.MatrixWidth = %d; must be > 0",
            c.Hardware.MatrixWidth)
        return errors.New(e)
    }
    if c.Hardware.MatrixHeight < 1 {
        e := fmt.Sprintf("Hardware.MatrixHeight = %d; must be > 0",
            c.Hardware.MatrixHeight)
        return errors.New(e)
    }
    if !contains(hardwareMappings, c.Hardware.Mapping) {
        e := fmt.Sprintf("Hardware.Mapping = %s; must be one of: %s",
            c.Hardware.Mapping, strings.Join(hardwareMappings, ", "))
        return errors.New(e)
    }

    c.loadSchedules(f)

    return nil
}

func (c *Config) getScheduleState(nd string, nt string, state bool) bool {
    ts, ok := c.schedules[nd]
    if !ok {
        return state
    }

    debugLog("schedule: day = %s, time = %s\n", nd, nt)

    now := toTime(nt, state)
    var tm Time

    for _, t := range ts {
        c := now.Compare(t)
        if c == 0 || c == 1 {
            state = t.state
            tm = t
        }
    }

    if debug && state != now.state {
        b := "off"
        if state {
            b = "on"
        }
        log.Printf("schedule: %s @ %s\n", b, tm)
    }

    return state
}
