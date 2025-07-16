# fad

`fad` is a command-line tool which (F)inds recently (A)ctive (D)irectories. The theory
being that your interest in recently active directories is more than just a passing
fad... ahem.

`fad` recursively searches each path to find directories with the most recent **activity
date**.  The **activity date** is derived from the date-time-modified of the most recently
modified entry *within* each directory.

`fad` differs from "`find -mtime`" and "`find -newer`" in that it analyses activity
*within* each directory to confer an **activity date** on the parent as opposed to only
comparing the directory's date-time-modified against a fixed age or file.

The output lists directories in **activity date** order starting with the most recently
active first. By way of example, if `fad` scans the following directory structure on
02 Jun at around 13:30:

| Path | Date Time | Modified |
| :---------- | ------ | ----------- |
| Projects/nextgen/README.md | 01 Jun | 01:00 |
| Projects/nextgen/main.go | 28 May | 09:02 |
| Projects/parago/main.go | 10 May | 15:19 |
| Projects/parago/capture.go | 25 May | 23:21 |
| Projects/sparsify/main.cc | 13 Apr | 14:45 |
| Projects/sparsify/threads.cc | 02 Jun | 12:18 |

it produces the following output:

```cat
1h:f:Projects/sparsify/threads.cc
5D:f:Projects/nextgen/main.go
1W:f:Projects/parago/capture.go
```

which shows that the `sparsify` project is the most recently active due to `threads.cc`
being modified 1 hour ago and `nextgen` is the next most recently active due to `main.go`
being modified 5 days ago. To reduce clutter and remain directory focussed, `fad` only ever
lists a directory once regardless of its contents. The manpage describes the output
format.

Typical questions `fad` is designed to answer are:

| Question | Invocation
| :---- | :---------
| Which project have I been working on recently? | `fad` ~/Projects
| What `syslog` directories have been written to recently? | `fad` /var/log
| Have any configuration directories just changed? | `fad` -age 5m /etc /usr/local/etc /opt/etc
| My boss wants to know what I worked on last week - help! | `fad` -age 1W $HOME

### Project Status

[![Build Status](https://github.com/markdingo/fad/actions/workflows/go.yml/badge.svg)](https://github.com/markdingo/fad/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/markdingo/fad/branch/main/graph/badge.svg)](https://codecov.io/gh/markdingo/fad)
[![Go Report Card](https://goreportcard.com/badge/github.com/markdingo/fad)](https://goreportcard.com/report/github.com/markdingo/fad)

`fad` is known to run on FreeBSD, macOS, Linux and Windows and requires go 1.19 or later.

## Installation

`fad` can be installed with "`go install`" if you only wish to install the executable in
your home directory or you can use the more traditional "`make install`" method if you want
to install the executable and manpage in `/usr/local`.

### `go install`

```bash
go install github.com/markdingo/fad@latest
```
for the latest official release, or if you're after the leading edge:

```bash
go install github.com/markdingo/fad@main
```

`go install` performs all the downloading, building and installing with the executable
ending up in either `$GOPATH/bin` or `$HOME/go/bin`, depending on your setup.

### `make install`

The main reason to prefer the traditional install method is so that the executable and
manpage are installed in `/usr/local/`. With this method you need to clone the repo and
run `make`:

```bash
git clone https://github.com/markdingo/fad.git
cd fad
make all
sudo make install
```

A good first test of `fad` is to ask it to find your recently active installation
directories with:


```bash
fad $HOME /usr/local
```

All being well `fad` should display the directory containing iself at or near the top of
the list.

### Community

If you have any problems using `fad` or suggestions on how it can do a better job,
don't hesitate to create an [issue](https://github.com/markdingo/fad/issues) on
the project home page. This package can only improve with your feedback.

### Copyright and License

`fad` is Copyright :copyright: 2025 Mark Delany. This software is licensed under the
BSD 2-Clause "Simplified" License.
