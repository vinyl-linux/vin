# vin

[![Coverage Status](https://coveralls.io/repos/github/vinyl-linux/vin/badge.svg?branch=main)](https://coveralls.io/github/vinyl-linux/vin?branch=main)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=vinyl-linux_vin&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=vinyl-linux_vin)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=vinyl-linux_vin&metric=security_rating)](https://sonarcloud.io/dashboard?id=vinyl-linux_vin)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=vinyl-linux_vin&metric=sqale_index)](https://sonarcloud.io/dashboard?id=vinyl-linux_vin)

`vin` is the Vinyl Linux package manager. It is backed by git repo, runs as a client/server pair, supports arbitrary extra package repositories, has portage inspired use flags, and supports version ranges for dependencies.


## Rationale

There are a number of issues with many package managers. They can be slow to start as they load/reload config, they don't handle queing of operations (relying, instead, on fragile lock files), they have version constraints which leave maintainers scrabbling to update in times of vulnerability mitigation, they use centralised package lists which need all kinds of extra tooling to grant access to, and then extra governance to update and test.

The list goes on.

`vin` is designed to either solve or even side-step some of these pitfalls

1. It runs as a client/server; this allows us to keep config and manifests in memory, rather than reading each time
1. This paradigm also allows us to accept and enqueue operations
1. Packages list dependencies as constraint ranges, powered by https://github.com/hashicorp/go-version
1. Packages are stored and distributed as git repos, allowing for ease of maintainance, change tracking, security, and more
1. Packages can come from any number of directories through the environment variable `VIN_PATH` - this allows easy testing of custom packages, or even simple distribution of third parties.
1. This approach also allows users to patch package manifests, or even just peek at them

It also uses blazing fast, nice and modern tooling. Downloads, for example, are checkum'd with https://github.com/BLAKE3-team/BLAKE3, which provides all kinds of lovely speed and security.

## Special cases

Not all packages download a tarball, perform some patch/build/installation steps, and complete. These are detailed here

### Meta packages

For packages whose only job is to include other packages as dependencies, such as [vinyl-0.1.0](https://github.com/vinyl-linux/vin-packages-stable/blob/main/vinyl/0.1.0/manifest.toml) from the stable repository.

In this situation, simply omit tarball and checksum, setting `meta = true`.

```toml
provides = "vinyl"
version = "0.1.0"
meta = true

[profiles]
[profiles.default]
deps = [
     ["vinit", "=0.5.1"],
     ["vin", "=0.9.0"],
     ["linux-utils", "=0.1.1"],
     ["vc", "=0.2.1"],
]
```

### Service directory packages

For packages which only include service directories for [`vinit`](https://github.com/vinyl-linux/vinit), use the tarball [https://github.com/vinyl-linux/vin/releases/download/0.10.1/empty.tar.gz](https://github.com/vinyl-linux/vin/releases/download/0.10.1/empty.tar/gz), which has the blake3 sum `3e767181b1a035d296bf393e35c65441bb0158141a0cfb51cf389b60ab01e8be`.

This is an empty tarball of about 4kb, so should download and process pretty quickly.

If this becomes too much of a faff later we can change the manifest spec to allow skipping tarballs.

## Testing

This repo uses github actions to test both the main branch and changes to the main branch. To run tests manually, run

```bash
$ go test -v ./...
```

This is the standard go test process, and should not contain any surprises.

We use sonarqube to keep an eye on code quality/ security scanning. This can be seen: https://sonarcloud.io/dashboard?id=vinyl-linux_vin

### Testing gRPC endpoints

This project contains client binaries, and but can also be accessed with grpcurl:

```bash
$ grpcurl -d '{"pkg":"sample-app"}' -plaintext -unix ${VIN_SOCK_ADDR} server.Vin.Install
```
