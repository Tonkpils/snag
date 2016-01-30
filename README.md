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

Once snag is installed, use:

```sh
snag init
```

This will generate the snag file `.snag.yml`.
Here is a sample of the snag file:

```yaml
verbose: true
ignore:
  - .git
  - "**.ext"
  - "foo/**/bar.sh"
build:
  - echo "hello world"
  - go test
```

Snag works by reading the snag file allowing you to configure what and how
commands will be executed.
The file **must** reside in the same directory that you want to watch.

By default, snag will watch all files/folders within the current directory recursively.
The ignore section will tell snag to ignore any changes that happen
to the directories/files listed. The ignore section uses the same pattern matching
that [gitignore](https://www.kernel.org/pub/software/scm/git/docs/gitignore.html) uses.

The build section of the file will be executed when any file is created, deleted, or modified.

Once configured, use:

```
snag
```

From a project with a snag file and develop away!

### Quick Use

If you find yourself working on a project that does not contain a snag file and
still want to use snag, you can use flags to specify commands to run.

The `-c` flag allows specifying a command just like the snag file and can
be defined more than once for multiple commands. The order of the commands
depends on the order of the flag.

```sh
snag -c "echo snag world" -c "echo rocks"
```

will output

```sh
|Passed     | echo snag world
|Passed     | echo rocks
```

The `-v` flag enables verbose output. It will also override the `verbose`
option form the snag file if it is defined to false.

**NOTE**: using the `-c` flag will skip reading a snag file even if it
exists in the current working directory.

### Environment Variables

You can access your shell's environment variables by using `$$`.

```yaml
build:
  - echo $$MY_VAR
  - rm -rf $$OUTPUT_DIR
```

## Caveats

### Endless build loops

Snag will run your configured commands when **ANY** file, not ignored,
is modifed in your current directory.
If your commands generate any files within the watched directory,
you must add them to the `ignore` section in your
snag file to avoid an endless build loop.

### Trouble running shell scripts

In order to run shell scripts, your must have a [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) in it. If you are trying to run a script without a
shebang, you will need to specify the shell it should run in.

i.e.

Running a script with a shebang

```yaml
build:
  - ./my-script
```

Running a script **without** a shebang

```yaml
build:
  - bash my-script
```

### Ignore Pattern Matching

If you want to use asterisks in the ignore section of your snag file,
you need to make sure to wrap them in quotes or you may run into an
error like:

```
$ snag
2015/10/24 19:39:40 Could not parse yml file. yaml: line 6: did not find expected alphabetic or numeric character
```

## Known Issues

* `open /dev/null: too many open files`

You may experience this error if you're running on OSX. You may need to bump
the maximum number of open file on your machine. You can refer to [this](http://krypted.com/mac-os-x/maximum-files-in-mac-os-x/)
article for more information on the max files on OSX and [this](http://superuser.com/questions/433746/is-there-a-fix-for-the-too-many-open-files-in-system-error-on-os-x-10-7-1) superuser post for a solution
