# Snag [![Build Status](https://travis-ci.org/Tonkpils/snag.svg?branch=wip)](https://travis-ci.org/Tonkpils/snag) [![Coverage Status](https://coveralls.io/repos/Tonkpils/snag/badge.svg?branch=coverage&service=github)](https://coveralls.io/github/Tonkpils/snag?branch=coverage)

## Installation

If you have [go](http://golang.org/) installed and want to install
the latest and greatest you can run:

```go
$ go get github.com/Tonkpils/snag
```

If you do not have go installed on your machine, you can checkout
the [releases](https://github.com/Tonkpils/snag/releases) section to
download the binary for your platform.

## Usage

### Running

Using snag is as easy and changing into your go projects directory
and running `snag`. Be default it will run the following commands

```bash
go build ./...
go vet ./...
go test ./...
```

### Using Arguments

You can pass arguments to snag to specify what packages to run against and
what flags to pass to the individual commands.

```bash
Usage of snag:
  -build=[]: comma delimited list of arguments given to the build command
  -build-tool="go": build tool used to run commands. (Godeps, GB and Go are the only ones currently supported)
  -packages="./...": comma delimited list of packages to run commands on
  -test=[]: comma delimited list of arguments given to the test command
  -vet=[]: comma delimited list of arguments given to the vet command
```

---

![](http://i.imgur.com/Vh7daqm.gif)