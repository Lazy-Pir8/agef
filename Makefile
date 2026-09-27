PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

.PHONY: all build rpm deb packages test clean install uninstall

all: build

build:
	@echo "Building agef binary..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/agef ./cmd/agef

rpm: build
	@echo "Building RPM package for agef..."
	@./packaging/build-rpm.sh

deb: build
	@echo "Building DEB package for agef..."
	@./packaging/build-deb.sh

packages: rpm deb
	@echo "All Linux packages built in dist/:"
	@ls -lh dist/

test:
	@echo "Running tests..."
	go test -v ./...

install: build
	@echo "Installing to $(BINDIR)..."
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 bin/agef $(DESTDIR)$(BINDIR)/agef
	@echo "Installed agef to $(BINDIR)"

uninstall:
	@echo "Uninstalling from $(BINDIR)..."
	rm -f $(DESTDIR)$(BINDIR)/agef

clean:
	rm -rf bin build dist
