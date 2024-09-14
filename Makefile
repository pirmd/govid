.POSIX:
PREFIX  ?= /var/www
CGIDIR  ?= ${PREFIX}/cgi-bin
HTDOCS  ?= ${PREFIX}/htdocs/govid

GO             = go
LDFLAGS_STATIC = -linkmode 'external' -extldflags '-static'
LDFLAGS        = ${LDFLAGS_STATIC} -w -s

BIN       = govid
SRC_MOD   = go.mod
SRC       != ${GO} list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...
SRC_TEST  != ${GO} list -f '{{range .TestGoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...
SRC_EMBED != ${GO} list -f '{{range .EmbedFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...

VIJS_PRJ = vi.js
VIJS_SRC != ls ${VIJS_PRJ}/src/*.ts
VIJS     = ${VIJS_PRJ}/vi.js
VIJSMIN  = ${VIJS_PRJ}/vi.min.js

ASSETS != ls htdocs/*
ASSETS += ${VIJSMIN}

.PHONY: all clean dev-dep 

all: audit ${BIN} ${ASSETS}

${BIN}: ${SRC} ${SRC_MOD} ${SRC_EMBED} ${SRC_TEST}
	${GO} fmt ./...
	${GO} mod tidy -v
	${GO} mod verify
	${GO} test -vet=all ./...
	${GO} build -ldflags "${LDFLAGS}" -o $@
	@grep -xq "$@" .gitignore || echo $@ >> .gitignore

audit: ${SRC} ${SRC_MOD}
	${GO} run honnef.co/go/tools/cmd/staticcheck@latest ./...
	${GO} run github.com/kisielk/errcheck@latest ./...
	${GO} run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
	${GO} run golang.org/x/vuln/cmd/govulncheck@latest ./...
.PHONY: audit

${VIJS}: ${VIJS_SRC}
	cd ${VIJS_PRJ} && npm run -s fmt
	cd ${VIJS_PRJ} && npm run -s check
	cd ${VIJS_PRJ} && npm run -s build

${VIJSMIN}: ${VIJS}
	cd ${VIJS_PRJ} && npm run -s minify

install: ${BIN} ${ASSETS}
	@echo "* Install ${BIN} to ${DESTDIR}${CGIDIR}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${CGIDIR}
	${INSTALL} -o root -g daemon -m 0755 ${BIN} ${DESTDIR}${CGIDIR}

	@echo "* Install static assets to ${DESTDIR}${HTDOCS}"
	${INSTALL} -d -o root -g daemon -m 0755 ${DESTDIR}${HTDOCS}
	${INSTALL} -o root -g daemon -m 0644 ${ASSETS} ${DESTDIR}${HTDOCS}

clean:
	go clean
	-rm -f ${BIN} ${VIJS} ${VIJSMIN}

dev-dep:
	@echo "* Install Typescript developpement environnement"
	cd ${VIJS_PRJ} && npm clean-install
