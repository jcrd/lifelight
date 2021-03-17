LIBDIR = lib/rpi-rgb-led-matrix/lib
LIB = $(LIBDIR)/librgbmatrix.so.1

all: $(LIB)

install: $(LIB)
	go install -v ./...

$(LIB):
	$(MAKE) -C $(LIBDIR)

clean:
	$(MAKE) -C $(LIBDIR) clean

.PHONY: all install clean
