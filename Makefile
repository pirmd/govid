.POSIX:
PREFIX  ?= /var/www
CGIDIR  ?= ${PREFIX}/cgi-bin
HTDOCS  ?= ${PREFIX}/htdocs/govid
NOTESDIR ?= ${PREFIX}/notes

GO = go

BIN = govid
SRC = govid.go
ASSETS = htdocs/index.html htdocs/vi-minimal.js

.PHONY: all clean install run test audit

all: test ${BIN}

${BIN}: ${SRC}
	${GO} build -ldflags "-s -w" -o $@ $<

install: ${BIN} ${ASSETS}
	@echo "* Installing ${BIN} to ${DESTDIR}${CGIDIR}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${CGIDIR}
	${INSTALL} -o root -g daemon -m 0755 ${BIN} ${DESTDIR}${CGIDIR}

	@echo "* Installing static assets to ${DESTDIR}${HTDOCS}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${HTDOCS}
	${INSTALL} -o root -g daemon -m 0644 ${ASSETS} ${DESTDIR}${HTDOCS}

	@echo "* Creating notes directory at ${DESTDIR}${NOTESDIR}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${NOTESDIR}

clean:
	rm -f ${BIN}

run: ${BIN}
	@echo "Starting govid in standalone mode (Ctrl+C to stop)"
	@echo "Access at: http://localhost:8080"
	GOVID_DIR=./notes ./${BIN}

test:
	${GO} test -v .

audit:
	${GO} run honnef.co/go/tools/cmd/staticcheck@latest ./
	${GO} run github.com/kisielk/errcheck@latest ./
	${GO} run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./
	${GO} run golang.org/x/vuln/cmd/govulncheck@latest ./
