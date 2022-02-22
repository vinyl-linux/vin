PREFIX ?= ""
BINDIR ?= "$(PREFIX)/usr/bin"
ETCDIR ?= "$(PREFIX)/etc/vinyl"
CACHEDIR ?= "$(PREFIX)/var/cache/vinyl/vin"
PKGDIR ?= "$(ETCDIR)/pkg"
SRVDIR ?= "$(PREFIX)/etc/vinit/services/vind"

OWNER ?= "root"

DIRS := $(BINDIR)     \
	$(ETCDIR)     \
	$(CACHEDIR)   \
	$(PKGDIR)     \
	$(SRVDIR)     \

BINARIES := $(BINDIR)/vind \
	    $(BINDIR)/vin

CONFIGS := $(ETCDIR)/vin.toml

SERVICES := $(SRVDIR)/wd           \
	    $(SRVDIR)/environment  \
	    $(SRVDIR)/.config.toml \
	    $(SRVDIR)/bin

BUILT_ON := $(shell date --rfc-3339=seconds | sed 's/ /T/')
BUILT_BY := $(shell whoami)
BUILD_REF := $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)

.PHONY: default
default: vind vin

$(DIRS):
	mkdir -vp $@

dirs: $(DIRS)

server/:
	mkdir -p $@

server/install.pb.go server/server.pb.go server/server_grpc.pb.go: **/*.proto | server/
	protoc --proto_path=proto --go_out=server --go-grpc_out=server --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative install.proto server.proto

vind: *.go server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.ref=$(BUILD_REF) -X main.buildUser=$(BUILT_BY) -X main.builtOn=$(BUILT_ON)" -trimpath -o $@

vin: pkg = "github.com/vinyl-linux/vin/client/cmd"
vin: client/*.go client/**/*.go server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	(cd client && CGO_ENABLED=0 go build -ldflags="-s -w -X $(pkg).Ref=$(BUILD_REF) -X $(pkg).BuildUser=$(BUILT_BY) -X $(pkg).BuiltOn=$(BUILT_ON)" -trimpath -o ../$@)

binInstallCmd     ?= install -m 0700 -o $(OWNER)
regInstallCmd     ?= install -m 0600 -o $(OWNER)

.PHONY: install
install: dirs $(BINARIES) $(CONFIGS)

.PHONY: install-service
install-service: $(SERVICES)

$(BINDIR)/%: % | $(BINDIR)
	$(binInstallCmd) $< $@

$(ETCDIR)/vin.toml: $(BINDIR)/vin | $(ETCDIR)
	-mv $@ $@.bak
	$< advise > $@

$(SRVDIR)/wd $(SRVDIR)/bin:
	ln -svf $(PREFIX)$(shell readlink service/$(notdir $@)) $@

$(SRVDIR)/%: service/% | $(SRVDIR)
	$(regInstallCmd) $< $@
