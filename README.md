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
