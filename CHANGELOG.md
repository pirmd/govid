# Changelog
## [0.3.1] - 2024-09-15
- Fix CGI DOCUMENT_URI usage by removing any SCRIPT_NAME prefix.
- Fix browser's template that improperly build links to resources.
- Fix vi.js "J" command improper behavior.
- Fix misbehavior of vi.js 'dd' command.
- Fix non working vi.js ':wq' command.
- Fix misbehavior of vi.js 'Tab' in INSERT mode.
- Fix misbehavior of vi.js 'O' command.
- Filter out entries with invalid filename when listing folder content.
- Normalize new lines to unix representation when saving node content.
- Add to vi.js ':e' command support to open notes relative to current path.
- Change default assets installation location to /var/www/htdocs/govid instead
  of /var/www/htdocs.
- Add GOVID_NOTESDIR environment variable to configure the notes folder that
  govid shall serve.
- Add GOVID_URL_PREFIX environment variable to configure the URL PREFIX.
- Add support for deadkey detection to vi.js. Introduce '^' command to move to
  first word of the current line.

## [0.3.0] - 2023-02-03
- switch from a standalone web-app to a simple CGI app.
- introduce home-brewed vi.js replacing jsvim.js by [Jakub
  Mikians](https://github.com/jakub-m) for fun, closer 'vi' look and feel, new
  minor features and probably more new bugs to hunt.
- add support to browse directories.
- add verification that requested file is not an hidden file or does not belong
  to an hidden folder.
- add a navigation bar.

## [0.2.0] - 2022-10-15
- change the way to specify folder that contain notes from command-line.
- modify API for an (hopefully) cleaner edit/save access.
- add support for basic authentication.

## [0.1.0] - 2022-09-10
- creation
- minimal set of features allowing basic interaction with plain-text files from
  the web-app. This version is more a proof of concept than anything else, it
  notably does not implement basic security approach for web-facing apps.


[modeline]: # ( vim: set fenc=utf-8 spell spl=en: )
