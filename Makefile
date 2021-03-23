rgbmatrix = rpi-rgb-led-matrix
LIBDIR = vendor/github.com/jcrd/go-$(rgbmatrix)
LIB = $(LIBDIR)/lib/$(rgbmatrix)/lib/librgbmatrix.so.1

SRC = life/life.go life/config.go life/log.go life/debug_log.go
SRC_TEST = life/life_test.go life/config_test.go

lifelight: main.go $(SRC) $(LIB)
	go build -o lifelight $<

debug: main.go $(SRC) $(LIB)
	go build -tags debug -o lifelight $<

$(LIB):
	$(MAKE) -C $(LIBDIR)

run: lifelight
	sudo ./lifelight

test: $(SRC) $(SRC_TEST)
	cd life && go test

clean:
	$(MAKE) -C $(LIBDIR) clean

.PHONY: debug run test clean
