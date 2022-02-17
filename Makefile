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
	CGO_ENABLED=0 go build -o vind

vin: client/*.go client/**/*.go server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	(cd client && CGO_ENABLED=0 go build -o ../vin)

installCmd     ?= install -m 0750 -o $(OWNER)

.PHONY: install
install: dirs $(BINARIES) $(CONFIGS) $(SERVICES)

$(BINDIR)/%: % | $(BINDIR)
	$(installCmd) $< $@

$(ETCDIR)/vin.toml: $(BINDIR)/vin | $(ETCDIR)
	-mv $@ $@.bak
	$< advise > $@

$(SRVDIR)/%: service/% | $(SRVDIR)
	$(installCmd) $< $@
