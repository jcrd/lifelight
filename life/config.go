package life

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/go-ini/ini"
)

var colorPalettes = []string{
	"happy",
	"soft",
	"warm",
}

var hardwareMappings = []string{
	"regular",
	"adafruit-hat",
	"adafruit-hat-pwm",
	"compute-module",
}

type Time struct {
	state bool
	hh    int
	mm    int
}

type Color struct {
	Scheme        []string
	Palettes      []string
	ScheduleRegen bool
}

type Hardware struct {
	MatrixWidth  int
	MatrixHeight int
	Mapping      string
}

type Config struct {
	TicksPerSecond          int
	SeedThreshold           float32
	SeedThresholdDecay      float32
	SeedThresholdDecayTicks int
	SeedCooldownTicks       int
	Schedule                bool

	Color
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

func NewConfig() *Config {
	c := &Config{
		TicksPerSecond:          12,
		SeedThreshold:           0.5,
		SeedThresholdDecay:      0.05,
		SeedThresholdDecayTicks: 5,
		SeedCooldownTicks:       2,
		Schedule:                true,
		Color: Color{
			Palettes:      colorPalettes,
			ScheduleRegen: true,
		},
		Hardware: Hardware{
			MatrixWidth:  32,
			MatrixHeight: 32,
			Mapping:      "adafruit-hat",
		},
	}
	c.schedules = make(map[string][]Time)

	return c
}

func (c *Config) loadSchedules(f *ini.File) {
sections:
	for _, section := range f.ChildSections("Schedule") {
		var times [2]string

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
			if k, err := section.GetKey(n); err == nil {
				times[i] = k.Value()
			}
		}

		for i, str := range times {
			if str == "" {
				continue
			}
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

		logger.log("config", "%s: on %s, off %s (%s)\n",
			name, times[1], times[0], strings.Join(days, ", "))
	}
}

func (c *Config) Load(path string, mustExist bool) error {
	if _, err := os.Stat(path); err != nil {
		if mustExist {
			return err
		} else {
			return nil
		}
	}

	logger.log("config", "Loading file '%s'...\n", path)

	f, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, path)
	if err != nil {
		return err
	}

	if err = f.MapTo(c); err != nil {
		return fmt.Errorf("Failed to parse file '%s': %v", path, err)
	}

	if c.TicksPerSecond < 1 {
		return fmt.Errorf("TicksPerSecond = %d; must be > 0",
			c.TicksPerSecond)
	}
	if c.SeedThreshold > 1.0 || c.SeedThreshold < 0.0 {
		return fmt.Errorf("SeedThreshold = %f; must be in range [0.0, 1.0]",
			c.SeedThreshold)
	}
	if c.SeedThresholdDecay < 0.0 {
		return fmt.Errorf("SeedThresholdDecay = %f; must be positive",
			c.SeedThresholdDecay)
	}
	if c.SeedThresholdDecayTicks < 0 {
		return fmt.Errorf("SeedThresholdDecayTicks = %d; must be positive",
			c.SeedThresholdDecayTicks)
	}
	if c.SeedCooldownTicks < 0 {
		return fmt.Errorf("SeedCooldownTicks = %d; must be positive",
			c.SeedCooldownTicks)
	}

	n := len(c.Color.Scheme)
	if n > 0 && n < 4 {
		return fmt.Errorf("Color.Scheme length = %d; must be 4", n)
	}

	for _, p := range c.Color.Palettes {
		if !contains(colorPalettes, p) {
			return fmt.Errorf("Color.Palettes contains %s; must be one of: %s",
				c.Color.Palettes, strings.Join(colorPalettes, ", "))
		}
	}

	if c.Hardware.MatrixWidth < 1 {
		return fmt.Errorf("Hardware.MatrixWidth = %d; must be > 0",
			c.Hardware.MatrixWidth)
	}
	if c.Hardware.MatrixHeight < 1 {
		return fmt.Errorf("Hardware.MatrixHeight = %d; must be > 0",
			c.Hardware.MatrixHeight)
	}
	if !contains(hardwareMappings, c.Hardware.Mapping) {
		return fmt.Errorf("Hardware.Mapping = %s; must be one of: %s",
			c.Hardware.Mapping, strings.Join(hardwareMappings, ", "))
	}

	c.loadSchedules(f)

	return nil
}

func (c *Config) HasSchedule(d string) bool {
	_, ok := c.schedules[d]
	return ok
}

func (c *Config) NumSchedules() int {
	return len(c.schedules)
}

func (c *Config) GetScheduleState(nd string, nt string, state bool) bool {
	ts, ok := c.schedules[nd]
	if !ok {
		return state
	}

	logger.log("schedule", "day = %s, time = %s\n", nd, nt)

	now := toTime(nt, state)
	var tm Time

	for _, t := range ts {
		if now.Compare(t) > -1 {
			state = t.state
			tm = t
		}
	}

	if state != now.state {
		logger.log("schedule", "on: %t @ %s\n", state, tm)
	}

	return state
}
