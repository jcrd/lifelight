// +build !debug

package life

type DummyLogger struct {
}

func (lg DummyLogger) init(domains string) {
}

func (lg DummyLogger) log(domain, format string, v ...interface{}) {
}

func init() {
	logger = DummyLogger{}
}
