# RVID - Remote VI Daemon

[![Go Reference](https://pkg.go.dev/badge/github.com/pirmd/rvid.svg)](https://pkg.go.dev/github.com/pirmd/rvid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/rvi)](https://goreportcard.com/report/github.com/pirmd/rvid)

`rvid` is a small web-app to remotely edit a bunch of text files in a
as-close-as-possible vi fashion. It aims mainly at taking/reading quick notes
in environment I'm not in control of (i.e. no ssh to my cloud server, or no vi)
using a simple browser.

Compared to already existing full-features note-taking app, `rvi` is really
basic, build for a personal note taking perspective without any
bell-and-whistles, trying to offer an as quick and simple way to quickly take
notes (open whatever browser you find, connect to your server pointing to the
file you'll like to edit, authenticate, edit it).

## INSTALLATION
With golang binary installed on your system, you just need to run:
̀``shell
go install github.com/pirmd/rvid
```

## USAGE
Usage can be obtained from `rvid`'s command line by running:
``` shell
rvid -help
```

Run rvid by
``` shell
./rvid $HOME/mynotes
```

then visiting http://localhost:8080/MyNewNote.txt should
bring you to a vi-like text editing form where you can
input text. Once done, clicking on 'Save' will create
$HOME/mynotes/MyNewNote.txt with the content you have enter.

## API
`GET /{filename}`:: view/edit note located at {filename} path within the notes
directory `rvid` instance is serving.

`POST /{filename}`:: save note located at {filename} path within the notes
directory `rvid` instance is serving.

## CREDITS
`rvid` is using
[vim-in-textarea](https://github.com/jakub-m/vim-in-textarea) from
[Jakub Mikians](https://github.com/jakub-m) that offers a simple
and efficient way to interact with a textarea in a vim-like
fashion. Thanks to him!

## CONTRIBUTION
If you feel like to contribute, just follow github guidelines on
[forking](https://help.github.com/articles/fork-a-repo/) then [send a pull
request](https://help.github.com/articles/creating-a-pull-request/)


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )
