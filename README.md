# GOVID - Go VI Daemon

[![Go Reference](https://pkg.go.dev/badge/github.com/pirmd/govid.svg)](https://pkg.go.dev/github.com/pirmd/govid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/rvi)](https://goreportcard.com/report/github.com/pirmd/govid)

`govid` is a small web-app to remotely edit a bunch of text files in a
as-close-as-possible vi fashion. It aims mainly at taking/reading quick notes
in environment I'm not in control of (i.e. no ssh to my cloud server, or no vi)
using a simple browser.

Compared to already existing full-features note-taking app, `govid` is really
basic, build for a personal note taking perspective without any bell-and-whistles.
It tries to offer an as quick and simple way to quickly take notes (open
whatever browser you find, connect to your server pointing to the file you'll
like to edit, authenticate, edit it).

`govid` design is kept as minimal as possible, it notably does not take care of
TLS nor of setting proper headers with reasonably secured Content Security
Policy. The principle is to delegate it to battle-tested services that handle
that perfectly well, so `govid` is probably better run behind a proxy like
Openbsd's relayd(8).

Such approach has also the added benefit of making it relatively easy to run
`govid` in a chroot with dropped privileged of a dedicated user.

## INSTALLATION
With golang binary installed on your system, you just need to run:
̀``shell
go install github.com/pirmd/govid
```

## USAGE
Usage can be obtained from `govid`'s command line by running:
``` shell
govid -help
```

Run govid by
``` shell
./govid $HOME/mynotes
```

then visiting http://localhost:8888/MyNewNote.txt should bring you to a vi-like
text editing form where you can input text. Once done, clicking on 'Save' or
input ':w' in COMMAND mode will create $HOME/mynotes/MyNewNote.txt with the
content you have enter.

## API
`GET /{filename}`:: view/edit note located at {filename} path within the notes
directory `govid` instance is serving.

`POST /{filename}`:: save note located at {filename} path within the notes
directory `govid` instance is serving.

## DEPLOYEMENT

`govid` is better run behind a proxy like relayd(8), for example with a simple
relayd.conf(5):
``` shell
public_ipv4="WWW.XXX.YYY.ZZZ"
table <govid> { 127.0.0.1 }
govid_port="8888"

http protocol "https_reverse_proxy" {
    match header set "X-Client-IP" value "$REMOTE_ADDR:$REMOTE_PORT"
    match header set "X-Forwarded-For" value "$REMOTE_ADDR"
    match header set "X-Forwarded-By" value "$SERVER_ADDR:$SERVER_PORT"

    match response header set "Content-Security-Policy" value "default-src 'self'"
    match response header set "Referrer-Policy" value "no-referrer"
    match response header set "Strict-Transport-Security" value "max-age=15552000; includeSubDomains; preload"
    match response header set "X-Content-Type-Options" value "nosniff"
    match response header set "X-Frame-Options" value "SAMEORIGIN"
    match response header set "X-XSS-Protection" value "1; mode=block"
}

relay https {
    listen on $public_ipv4 port https tls
    protocol "https_reverse_proxy"
    forward to <govid> port $govid_port
}
```

Nota: above example expects that TLS keys and certificates are to be found in
/etc/ssl/WWW.XXX.YYY.ZZZ.crt and /etc/ssl/private/WWW.XXX.YYY.ZZZ.key.

You obvisouly can use whatever proxy you want.

I usually set-up a dedicated non-priviledge user for running `govid`:
``` shell
groupadd _govid
useradd -d /var/govid -c "Go VI Daemon" -g _govid -L daemon -s /sbin/nologin _govid
mkdir /var/govid && chown _govid:_govid /var/govid
```

Then add some credentials to limit access to your notes, for example:
``` shell
htpasswd /var/govid/htpasswd govid_user
```

Once done, you can run govid with `/usr/local/bin/govid -htpasswd /var/govid/htpasswd /var/govid/notes`
and access it pointing any browser to https://WWW.XXX.YYY.ZZZ/MyNotes.txt

At this point you are a step away from chrooting govid into /var/govid, for example on OpenBSD 7.1:
``` shell
mkdir -p /var/govid/usr/lib && cp /usr/lib/libc.so.96.1 /var/govid/usr/lib/
cp /usr/lib/libpthread.so.26.1 /var/govid/usr/lib/
mkdir -p /var/govid/usr/libexec && cp /usr/libexec/ld.so /var/govid/usr/libexec/
mkdir -p /var/govid/bin && cp /usr/local/bin/govid /var/govid/bin/

chroot -u _govid -g _govid /var/govid /bin/govid --address 127.0.0.1:8888 -htpasswd ./htpasswd ./notes >> ./log/access.log
```

## API
Supported request are:
+ `GET /{filename}`:: view/edit note located at {filename} path within the notes
directory `govid` instance is serving.

+ `POST /{filename}`:: save note located at {filename} path within the notes
directory `govid` instance is serving.

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
