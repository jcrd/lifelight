lifelight: main.go
	go build -o $@ $^

run: lifelight
	sudo ./lifelight

test: main_test.go main.go
	go test

.PHONY: run test
