package main

import (
    "log"
    "os"

    "github.com/go-ini/ini"
)

const configPath = "/etc/lifelight.ini"

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
}

func loadConfig() (c Config) {
    c = Config{
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

    path := configPath
    v, hasEnv := os.LookupEnv("LIFELIGHT_CONFIG")
    if hasEnv {
        path = v
    }

    if _, err := os.Stat(path); err != nil {
        if hasEnv {
            log.Printf("%v\n", err)
        }
        return c
    }

    debugLog("Loading config file...\n")

    if err := ini.MapTo(&c, path); err != nil {
        log.Printf("Failed to load config file %s: %v\n", path, err)
    }

    return c
}
