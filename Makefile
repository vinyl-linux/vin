PREFIX ?= ""
BINDIR ?= "$(PREFIX)/usr/bin"
ETCDIR ?= "$(PREFIX)/etc/vinyl"
CACHEDIR ?= "$(PREFIX)/var/cache/vinyl/vin"
PKGDIR ?= "$(ETCDIR)/pkg"
SRVDIR ?= "$(PREFIX)/etc/s6/sv/vind"

OWNER ?= "root"

DIRS := $(BINDIR)   \
	$(ETCDIR)   \
	$(CACHEDIR) \
	$(PKGDIR)   \
	$(SRVDIR)

.PHONY: default
default: vind vin

$(DIRS):
	mkdir -vp $@

dirs: $(DIRS)

server/:
	mkdir -p $@

server/install.pb.go server/server.pb.go server/server_grpc.pb.go: server/ **/*.proto
	protoc --proto_path=proto --go_out=server --go-grpc_out=server --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative install.proto server.proto

vind: *.go server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	go build -o vind

vin: client/*.go client/**/*.go server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	(cd client && go build -o ../vin)

installCmd     ?= install -m 0750 -o $(OWNER)
confInstallCmd ?= install -m 0640 -o $(OWNER)

.PHONY: install
install: dirs $(BINDIR)/vind $(BINDIR)/vin $(ETCDIR)/vin.toml $(SRVDIR)/run $(SRVDIR)/finish $(SRVDIR)/type $(SRVDIR)/conf

$(BINDIR)/vind: vind $(BINDIR)
	$(installCmd) $< $@

$(BINDIR)/vin: vin $(BINDIR)
	$(installCmd) $< $@

$(ETCDIR)/vin.toml: $(ETCDIR)
	$(confInstallCmd) /dev/null $@

$(SRVDIR)/%: service/% $(SRVDIR)
	$(installCmd) $< $@
