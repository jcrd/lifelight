PREFIX ?= /usr/local
BINPREFIX ?= $(PREFIX)/bin
ETCPREFIX ?= $(PREFIX)/etc
LIBPREFIX ?= $(PREFIX)/lib

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

install:
	mkdir -p $(DESTDIR)$(BINPREFIX)
	cp -p lifelight $(DESTDIR)$(BINPREFIX)
	mkdir -p $(DESTDIR)$(ETCPREFIX)
	cp -p lifelight.ini $(DESTDIR)$(ETCPREFIX)
	mkdir -p $(DESTDIR)$(LIBPREFIX)/systemd/system
	cp -p systemd/lifelight.service $(DESTDIR)$(LIBPREFIX)/systemd/system

uninstall:
	rm -f $(DESTDIR)$(BINPREFIX)/lifelight
	rm -f $(DESTDIR)$(ETCPREFIX)/lifelight.ini
	rm -f $(DESTDIR)$(LIBPREFIX)/systemd/system/lifelight.service

run: lifelight
	sudo ./lifelight

test: $(SRC) $(SRC_TEST)
	cd life && go test

clean:
	$(MAKE) -C $(LIBDIR) clean

.PHONY: debug install uninstall run test clean
