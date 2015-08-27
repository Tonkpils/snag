# Snag [![Build Status](https://travis-ci.org/Tonkpils/snag.svg?branch=wip)](https://travis-ci.org/Tonkpils/snag) [![Coverage Status](https://coveralls.io/repos/Tonkpils/snag/badge.svg?branch=coverage&service=github)](https://coveralls.io/github/Tonkpils/snag?branch=coverage)

An automatic build tool for all your needs

![](http://i.imgur.com/epcicvr.gif)

## Installation

### Releases

You can visit the [releases](https://github.com/Tonkpils/snag/releases) section to
download the binary for your platform.

## [Homebrew](http://brew.sh/)

We've got a formula in homebrew!

```bash
brew update && brew install snag
```

### Source

If you have [go](http://golang.org/) installed and want to install
the latest and greatest you can run:

```go
$ go get github.com/Tonkpils/snag
```

## Usage

Snag works by reading a yaml file named `.snag.yml`. It allows you to configure what snag will
run and what it should ignore. The file **must** reside in the same
directory that you want to watch.

Here is a sample of a `.snag.yml` file:

```yml
script:
  - echo "hello world"
  - go test
ignore:
  - .git
  - myfile.ext
verbose: true
```

By default, snag will watch all files/folders within the current directory recursively.
The ignore section will tell snag to ignore any changes that happen
in the `.git` directory and any changes that happen to the `myfile.ext` file.

The script section of the file will be executed when any file is created, deleted, or modified.

Simply run:

```
snag
```

From a project with a `.snag.yml` file and develop away!

## Caveats 

* Endless build loops

Snag will run your configured scripts if **ANY** file modifed in your current directory.
If you scripts generate any files, you should add them to the `ignore` section in your
`.snag.yml` to avoid an endless build loop.


## Known Issues

* `open /dev/null: too many open files`

You may experience this error if you're running on OSX. You may need to bump
the maximum number of open file on your machine. You can refer to [this](http://krypted.com/mac-os-x/maximum-files-in-mac-os-x/)
article for more information on the max files on OSX and [this](http://superuser.com/questions/433746/is-there-a-fix-for-the-too-many-open-files-in-system-error-on-os-x-10-7-1) superuser post for a solution
