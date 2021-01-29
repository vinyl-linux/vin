PREFIX ?= ""
BINDIR ?= "$(PREFIX)/usr/bin"
ETCDIR ?= "$(PREFIX)/etc/vinyl"
CACHEDIR ?= "$(PREFIX)/var/cache/vinyl/vin"
PKGDIR ?= "$(ETCDIR)/pkg"

OWNER ?= "root"

DIRS := $(BINDIR) \
	$(ETCDIR) \
	$(CACHEDIR) \
	$(PKGDIR)

.PHONY: default
default: vind

$(DIRS):
	mkdir -vp $@

dirs: $(DIRS)

server/:
	mkdir -p $@

server/install.pb.go server/server.pb.go server/server_grpc.pb.go: server/
	protoc --proto_path=proto --go_out=server --go-grpc_out=server --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative install.proto server.proto

vind: server/install.pb.go server/server.pb.go server/server_grpc.pb.go
	go build -o vind


installCmd     ?= install -m 0750 -o $(OWNER) -Cv
confInstallCmd ?= install -m 0640 -o $(OWNER) -Cv

.PHONY: install
install: dirs $(BINDIR)/vind $(ETCDIR)/config.toml

$(BINDIR)/vind: vind $(BINDIR)
	$(installCmd) $< $@

$(ETCDIR)/config.toml: $(ETCDIR)
	$(confInstallCmd) /dev/null $@
