lifelight: main.go
	go build -o $@ $^

test: main_test.go main.go
	go test

.PHONY: test
