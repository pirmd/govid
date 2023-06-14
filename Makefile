.POSIX:
PREFIX  ?= /var/www
CGIDIR  ?= ${PREFIX}/cgi-bin
HTDOCS  ?= ${PREFIX}/htdocs

GO             = go
LDFLAGS_STATIC = -linkmode 'external' -extldflags '-static'
LDFLAGS        = ${LDFLAGS_STATIC} -w -s

BIN       = govid
SRC_MOD   = go.mod
SRC      != go list -f '{{join .GoFiles " "}}'
SRC_TEST != go list -f '{{join .TestGoFiles " "}}'
SRC_TMPL != ls templates/*

VIJS_PRJ = vi.js
VIJS_SRC != ls ${VIJS_PRJ}/src/*.ts
VIJS     = ${VIJS_PRJ}/vi.js

ASSETS != ls htdocs/*
ASSETS += ${VIJS}

.PHONY: all clean dev-dep 

all: ${BIN} ${ASSETS}

${BIN}: ${SRC} ${SRC_MOD} ${SRC_TMPL} ${SRC_TEST}
	${GO} fmt ./...
	staticcheck ./...
	errcheck ./...
	gosec -quiet ./...
	${GO} test -vet=all ./...
	${GO} build -ldflags "${LDFLAGS}" -o $@

${VIJS}: ${VIJS_SRC}
	cd ${VIJS_PRJ} && npm run -s lint -- --quiet
	cd ${VIJS_PRJ} && npm run -s build

install: ${BIN} ${ASSETS}
	@echo "* Install ${BIN} to ${DESTDIR}${CGIDIR}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${CGIDIR}
	${INSTALL} -o root -g daemon -m 0755 ${BIN} ${DESTDIR}${CGIDIR}

	@echo "* Install static assets to ${DESTDIR}${HTDOCS}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${HTDOCS}
	${INSTALL} -o root -g daemon -m 0644 ${ASSETS} ${DESTDIR}${HTDOCS}

clean:
	go clean
	-rm -f ${BIN} ${VIJS}

dev-dep:
	@echo "* Install go verification tools"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/kisielk/errcheck@latest

	@echo "* Install Typescript developpement environnement"
	cd ${VIJS_PRJ} && npm clean-install
