rgbmatrix = rpi-rgb-led-matrix
LIBDIR = vendor/github.com/jcrd/go-$(rgbmatrix)
LIB = $(LIBDIR)/lib/$(rgbmatrix)/lib/librgbmatrix.so.1

SRC = main.go config.go
SRC_TEST = main_test.go config_test.go

lifelight: $(SRC) $(LIB)
	go build -o $@ $(SRC)

$(LIB):
	$(MAKE) -C $(LIBDIR)

run: lifelight
	sudo ./lifelight

run-debug: lifelight
	sudo LIFELIGHT_DEBUG=true ./lifelight

test: $(SRC) $(SRC_TEST)
	go test

clean:
	$(MAKE) -C $(LIBDIR) clean

.PHONY: run run-debug test clean
