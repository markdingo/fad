ROOT=/usr/local
# ROOT=$(HOME)
BINDIST=$(ROOT)/bin
MANDIST=$(ROOT)/man/man1
CMD=fad
MANPAGE=fad.1

GENERATED=version.go
CMDDEPENDS=$(GENERATED) *.go Makefile $(MANPAGE)

all: $(CMD)

$(CMD): $(CMDDEPENDS)
	go build -tags local

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
vet:	$(GENERATED)
	go vet -tags local ./...
	mandoc -Tlint $(MANPAGE); exit 0

.PHONY: install
install: $(BINDIST)/$(CMD) $(MANDIST)/$(MANPAGE)

.PHONY: test tests

# The "local" tag is my convention for any build or test requirements which can only be
# safely satisified on a local development system. This is largely used to turn off tests
# which fail on github because it presents a funky file system or some other non-standard
# behavior.
test tests: $(GENERATED)
	go test -tags local -race -v

$(BINDIST)/$(CMD): $(CMD) Makefile
	install -d -m u=rwx,go=rx $(BINDIST) # Ensure destination exists
	install -p -m a=rx $(CMD) $(BINDIST)
	@echo $(CMD) installed in $(BINDIST)
	@echo

$(MANDIST)/$(MANPAGE): $(MANPAGE) Makefile
	install -d -m u=rwx,go=rx $(MANDIST) # Ensure destination exists
	install -p -m a=r $(MANPAGE) $(MANDIST)
	@echo $(MANPAGE) installed in $(MANDIST)
	@echo

.PHONY: windows
windows: fad.exe
fad.exe: $(CMDDEPENDS)
	@echo 'Building for Windows amd64 (10 or higher)'
	@GOOS=windows GOARCH=amd64 go build -tags local
	@file $(CMD).exe
