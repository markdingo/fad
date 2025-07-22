# A simple Makefile for those who want to install fad the traditional way. The
# main benefits are that the manpage is installed and you get to control the
# installation location via "make install ROOT=*location*.
#
# This is primarily a BSD make file which also of course works with GNU
# make. However, GNU make generates a couple of specious warnings about the
# /usr/local target. Ignore these.
ROOT:=/usr/local	# make install ROOT=~ for a per-user install

# The "local" tag is my convention for any build or test requirements which can
# only be safely satisified on a local development system. This is largely used
# to turn off tests which fail on github because it presents a funky file system
# or some other non-standard behavior.
TAGS=-tags local

BINDIST=${ROOT}/bin
MANDIST=${ROOT}/man/man1
CMD=fad
MANPAGE=fad.1

GENERATED=version.go
CMDDEPENDS=${GENERATED} *.go Makefile ${MANPAGE}

all: ${CMD}

${CMD}: ${CMDDEPENDS}
	go build ${TAGS}

version.go:	generate_version.sh ChangeLog.md go.mod Makefile
	sh generate_version.sh ChangeLog.md version.go

.PHONY:	fmt
fmt:
	gofmt -s -w .

.PHONY: clean
clean:
	go clean
	rm -f fad.exe

.PHONY: vet
vet:	${GENERATED}
	go vet ${TAGS} ./...
	mandoc -Tlint ${MANPAGE}; exit 0

.PHONY: install
install: ${BINDIST}/${CMD} ${MANDIST}/${MANPAGE}

.PHONY: test tests

test tests: ${GENERATED}
	go test ${TAGS} -race -v

${BINDIST}/${CMD}: ${CMD} Makefile
	install -d -m u=rwx,go=rx ${BINDIST} # Ensure destination exists
	install -p -m a=rx ${CMD} ${BINDIST}
	@echo ${CMD} installed in ${BINDIST}
	@echo

${MANDIST}/${MANPAGE}: ${MANPAGE} Makefile
	install -d -m u=rwx,go=rx ${MANDIST} # Ensure destination exists
	install -p -m a=r ${MANPAGE} ${MANDIST}
	@echo ${MANPAGE} installed in ${MANDIST}
	@echo

.PHONY: windows
windows: fad.exe
fad.exe: ${CMDDEPENDS}
	@echo 'Building for Windows amd64 (10 or higher)'
	@GOOS=windows GOARCH=amd64 go build ${TAGS}
	@file ${CMD}.exe
