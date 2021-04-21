// +build debug

package life

import (
	"fmt"
	"log"
	"strings"
)

type DebugLogger struct {
	domains map[string]struct{}
}

func (lg *DebugLogger) init(domains string) {
	ds := make(map[string]struct{})
	for _, d := range strings.Split(domains, ",") {
		ds[d] = struct{}{}
	}
	if len(ds) > 0 {
		lg.domains = ds
	}
}

func (lg *DebugLogger) log(domain, format string, v ...interface{}) {
	if _, ok := lg.domains[domain]; ok {
		log.Printf(fmt.Sprintf("[%s] %s", domain, format), v...)
	}
}

func init() {
	logger = &DebugLogger{
		domains: map[string]struct{}{
			"config":   {},
			"schedule": {},
			"seed":     {},
		},
	}
}
