# GOVID - Go VI Daemon

[![Go Reference](https://pkg.go.dev/badge/github.com/pirmd/govid.svg)](https://pkg.go.dev/github.com/pirmd/govid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/rvi)](https://goreportcard.com/report/github.com/pirmd/govid)

`govid` is a CGI application to remotely edit a bunch of text files in a
as-close-as-possible vi fashion. It aims mainly at taking/reading quick notes
in environment I'm not in control of (i.e. no ssh to my cloud server, or no vi)
using a simple browser.

Compared to already existing full-features note-taking app, `govid` is really
basic, build for a personal note taking perspective with no bells nor whistles.
It tries to offer an "as quick and simple way" to quickly take notes (open
whatever browser you find, connect to your server pointing to the file you'll
like to edit, edit it).

`govid` design is intentionally kept as minimal as possible, leaving most of
the heavy work to whatever battle-tested http stack you want to use to deploy
it. 

`govid` is developed and used on OpenBSD, it is most probably going to operate
smoothly on any unix-like environment. Running `govid` on other plat-forms
like Windows might work through some features might not be properly supported
(like path validation logic).

## INSTALLATION AND DEPLOYMENT
To install `govid` CGI application, you can use:
̀``shell
make install
```

By default it will install:
- `govid` CGI application to ${CGIDIR} as well as its dependant libraries so
  that it can be run chrooted in ${PREFIX}.
- CSS and JS assets in ${HTDOCS}

where ${PREFIX} default to /var/www, ${CGIDIR} to ${PREFIX/cgi-bin} and
${HTDOCS} to ${PREFIX}/htdocs/govid. Each of these parameters can be altered
when invoquing
`make install`, for example:
```shell
make install PREFIX=my/prefered/www/location
```
## CONFIGURATION
`govid` will serve notes from the location contained in GODIR_NOTESDIR
environement variable or, if not set, from DOCUMENT_ROOT.

## API
Supported request are:
+ `GET /{filename}`:: view/edit file or folder located at {filename} path
  within the directory `govid` instance is serving.

+ `POST /{filename}`:: save file or folder located at {filename} path within
  the directory `govid` instance is serving.

{filename} corresponds to DOCUMENT_URI CGI environement variable without the
SCRIPT_NAME prefix (corresponds to PATH_INFO content). 

`govid` only accepts {filename} that lives inside govid's directory, it will
reject any path directives (like ../ or absolute path) that will try to save or
access files outside of this folder.
Addionally, {filename} pointing to hidden files (starting with '.') or files
living in an hidden folder are not accepted. Note that the logic implemented is
based on unix-like hidden files and is most certainly not going to operate well
on Windows.

If {filename} points to a non-existing file, it will be created once saving,
including any sub-folders. Files and sub-folders are created using the umask of
the user under which `govid` is running.

Requests for {filename} pointing to files that are believed not to be in
plaintext mime-type will be rejected. Likewise, POST request with a content
that does not look-like plaintext will be rejected.

Requests for accessing too big files or trying to save too big content will be
rejected.

## CONTRIBUTION
If you feel like to contribute, just follow github guidelines on
[forking](https://help.github.com/articles/fork-a-repo/) then [send a pull
request](https://help.github.com/articles/creating-a-pull-request/)


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )
