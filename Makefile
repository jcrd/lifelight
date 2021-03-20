rgbmatrix = rpi-rgb-led-matrix
LIBDIR = vendor/github.com/jcrd/go-$(rgbmatrix)
LIB = $(LIBDIR)/lib/$(rgbmatrix)/lib/librgbmatrix.so.1

lifelight: main.go $(LIB)
	go build -o $@ $<

$(LIB):
	$(MAKE) -C $(LIBDIR)

run: lifelight
	sudo ./lifelight

run-debug: lifelight
	sudo LIFELIGHT_DEBUG=true ./lifelight

test: main_test.go main.go
	go test

clean:
	$(MAKE) -C $(LIBDIR) clean

.PHONY: run run-debug test clean
