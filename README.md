# cirocket

Rocket powered task runner to assist delivering ci build missions

## Status

![Status](https://img.shields.io/badge/Status-ALPHA-red?style=for-the-badge)
[![Build Status](https://img.shields.io/circleci/build/gh/nehemming/cirocket/master?style=for-the-badge)](https://github.com/nehemming/cirocket) 
[![Release](https://img.shields.io/github/v/release/nehemming/cirocket.svg?style=for-the-badge)](https://github.com/nehemming/cirocket/releases/latest)
[![Coveralls](https://img.shields.io/coveralls/github/nehemming/cirocket?style=for-the-badge)](https://coveralls.io/github/nehemming/cirocket)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge)](/LICENSE)
[![GoReportCard](https://goreportcard.com/badge/github.com/nehemming/cirocket?test=0&style=for-the-badge)](https://goreportcard.com/report/github.com/nehemming/cirocket)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/goreleaser/goreleaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![Uses: cirocket](https://img.shields.io/badge/Uses-cirocket-orange?style=for-the-badge)](https://github.com/nehemming/cirocket)
[![Uses: GoReleaser](https://img.shields.io/badge/uses-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)


>This project is in its early alpha MVP stage.  It may be subject to change or redesign, but the hope is not to break any existing scripts using the tool.
>
>Documentation is limited.  Current sources include this README and the sample config file generated by `cirocket init mission`
>
>If you find and problem or want to contribute please see the contribution section below.

## Installation

The application can be installed via:

 * Pre-build [release binaries](https://github.com/nehemming/cirocket/releases)
 * Homebrew
 * As a [docker image](#docker)
 * Using the golang tool chain
 * From source by cloning this repository

### Homebrew

```sh
brew install nehemming/tap/cirocket
```

### Go tool chain

```sh
go install github.com/nehemming/cirocket@latest
```

### From source code

```sh
git clone https://github.com/nehemming/cirocket.git
cd cirocket
go install

```
# Basic usage

`cirocket` is a simple task runner application that uses a configuration file to specify the tasks.  

To create a sample config file, along with some basic documentation run:

```sh
cirocket init
```

This will create a `.cirocket.yml` file in the current working directory. 

Here is example [sample](https://github.com/nehemming/cirocket/blob/master/internal/cmd/initcoonfig.yml) 

Once edited the file can be run using

```sh
cirocket launch
```

Different mission files can be specified by adding to either command the `--mission <filename>` flag.

## <a name="docker"></a>Docker 

CI Rocket a basic docker image available in [packages](https://github.com/nehemming/cirocket/pkgs/container/cirocket).


To use the image, either pull it or include it in Dockerfile FROM statement.

```sh
docker pull ghcr.io/nehemming/cirocket:latest
```

The tool is best used with a mounted volume pointing at your project.

```
docker run --rm -ti -v /host/project:/project cirocket --dir /project init
```

>TIP:  The `--dir` flag switches to the supplied directory before running the tool.


## Contributing

We would welcome contributions to this project.  Please read our [CONTRIBUTION](https://github.com/nehemming/cirocket/blob/master/CONTRIBUTING.md) file for further details on how you can participate or report any issues.

## License

This software is licensed under the [Apache License](http://www.apache.org/licenses/). 