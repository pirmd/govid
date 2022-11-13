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
like to edit, authenticate, edit it).

`govid` design is intentionally kept as minimal as possible, leaving most of
the heavy work to whatever battle-tested http stack you want to use to deploy
it. 

## INSTALLATION AND DEPLOYMENT
To install `govid` CGI application, you can use:
̀``shell
make install
```

By default it will install:
- `govid` CGI application to ${CGIDIR} as well as its dependant libraries so
  that it can be run chrooted in ${WWWDIR}.
- CSS and JS assets in ${HTDOCS}

where $WWWDIR default to /var/www, $CGIDIR to $WWWDIR/cgi-bin and $HTDOCS to
$WWWDIR/htdocs. Each of these parameters can be altered when invoquing `make
install`, for example:
```shell
make install WWWDIR=my/prefered/www/location
```

## API
Supported request are:
+ `GET /{filename}`:: view/edit note located at {filename} path within the notes
directory `govid` instance is serving.

+ `POST /{filename}`:: save note located at {filename} path within the notes
directory `govid` instance is serving.

`govid` only accepts {filename} that lives inside govid's notes directory, it
will silently 'clean' any path directives (like ../ or absolute path) that will
try to save or access document outside of this folder.

If {filename} points to a non-existing note, it will be created once saving,
including any sub-folders. Notes and sub-folders are created using the umask of
the user under which `govid` is running.

Requests for {filename} pointing to files that are believed not to be in
plaintext mime-type will be rejected. Likewise, POST request with a content
that does not look-like plaintext will be rejected.

Requests for accessing too big files or trying to save too big content will be
rejected.

## CREDITS
`govid` is using
[vim-in-textarea](https://github.com/jakub-m/vim-in-textarea) from
[Jakub Mikians](https://github.com/jakub-m) that offers a simple
and efficient way to interact with a textarea in a vim-like
fashion. Thanks to him!

## CONTRIBUTION
If you feel like to contribute, just follow github guidelines on
[forking](https://help.github.com/articles/fork-a-repo/) then [send a pull
request](https://help.github.com/articles/creating-a-pull-request/)


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )
