.POSIX:
GO      = go
LDFLAGS = -w -s

BIN = govid
SRC = govid.go main.go
SRC_TMPL = templates/edit.html.gotmpl
SRC_MOD = go.mod
SRC_TEST = govid_test.go

INSTALL ?= install
WWWDIR  ?= /var/www
CGIDIR  ?= $(WWWDIR)/cgi-bin
HTDOCS  ?= $(WWWDIR)/htdocs
WWWUSR  ?= www
WWWGRP  ?= www

.PHONY: all clean install tools 

all: ${BIN}
${BIN}: ${SRC} ${SRC_MOD} ${SRC_TMPL} ${SRC_TEST}
	${GO} fmt ./...
	staticcheck ./...
	errcheck ./...
	gosec -quiet ./...
	${GO} test -vet=all ./...
	${GO} build -ldflags "${LDFLAGS}" -o $@

install: ${BIN}
	${INSTALL} -D -o ${WWWUSR} -g ${WWWGRP} -m 0750 ${BIN} ${DESTDIR}${CGIDIR}/${BIN}
	for d in `ldd ${BIN}|awk '$$5~1 {print $$7}'`; do \
		${INSTALL} -D -o root -g daemon -m 0644 $$d ${DESTDIR}${WWWDIR}$$d ; \
	done
	for d in `find htdocs/ -type f`; do \
		${INSTALL} -D -o ${WWWUSR} -g ${WWWGRP} -m 0640 $$d ${DESTDIR}${HTDOCS}$${d##htdocs} ; \
	done

clean:
	go clean
	rm -f ${BIN}

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/kisielk/errcheck@latest
