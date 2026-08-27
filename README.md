# GOVID - Minimalist Text Editor with Vi Mode

[![Go Reference](https://pkg.go.dev/badge/github.com/pirmd/govid.svg)](https://pkg.go.dev/github.com/pirmd/govid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/govid)](https://goreportcard.com/report/github.com/pirmd/govid)
[![License: BSD 2-Clause](https://img.shields.io/badge/License-BSD_2--Clause-blue.svg)](LICENSE)

`govid` is an **ultra-minimalist** text file editor accessible from a browser, featuring a **lightweight Vi emulator** (~340 lines of JS) for vi/vim users.

This version combines **extreme simplicity** (~100 lines of Go backend) with **basic Vi functionality** for a familiar experience.

## :rocket: Features

### Core
- Edit text files from a browser
- Auto-create directories
- Security validation (path traversal, hidden files, binary files)
- 1 MB file size limit
- CGI mode (Apache/Nginx) and standalone mode (dev on `:8080`)

### MiniVi Commands
| Command | Description |
|---------|-------------|
| `i` | Enter insert mode |
| `a` | Enter insert mode (after cursor) |
| `Esc` | Return to normal mode |
| `h` | Move cursor left |
| `j` | Move cursor down |
| `k` | Move cursor up |
| `l` | Move cursor right |
| `0` | Go to line start |
| `$` | Go to line end |
| `x` | Delete character under cursor |
| `dd` | Delete current line |
| `:w` | Save file |
| `:q` | Quit (close tab) |
| `:e filename` | Open another file |

## :gear: Installation

### Dependencies
- [Go 1.17+](https://golang.org/dl/) (for building)
- Web server (Apache, Nginx, Caddy) **or** none for standalone mode

### Build
```bash
make
```

### Deploy with Apache

1. Install Apache and enable CGI:
   ```bash
   sudo apt install apache2
   sudo a2enmod cgi
   ```

2. Configure Apache:
   ```apache
   # /etc/apache2/sites-available/govid.conf
   ScriptAlias /govid/ /var/www/cgi-bin/
   <Directory "/var/www/cgi-bin">
       AllowOverride None
       Options +ExecCGI
       Require all granted
       SetEnv GOVID_DIR "/var/www/notes"
       SetEnv GOVID_URL_PREFIX "/govid"
   </Directory>

   Alias /govid/static/ /var/www/htdocs/govid/
   <Directory "/var/www/htdocs/govid">
       Require all granted
   </Directory>
   ```

3. Install govid:
   ```bash
   sudo make install
   ```

4. Restart Apache:
   ```bash
   sudo systemctl restart apache2
   ```

5. Access at [http://your-server/govid/filename.txt](http://your-server/govid/filename.txt)

### Deploy with Nginx

1. Install Nginx and fcgiwrap:
   ```bash
   sudo apt install nginx fcgiwrap
   sudo systemctl start fcgiwrap.socket
   sudo systemctl enable fcgiwrap.socket
   ```

2. Configure Nginx:
   ```nginx
   # /etc/nginx/sites-available/govid
   server {
       listen 80;
       server_name your-server.com;

       location /govid/ {
           fastcgi_pass unix:/var/run/fcgiwrap.socket;
           include fastcgi_params;
           fastcgi_param SCRIPT_FILENAME /var/www/cgi-bin/govid;
           fastcgi_param GOVID_DIR /var/www/notes;
           fastcgi_param GOVID_URL_PREFIX /govid;
           fastcgi_param PATH_INFO $fastcgi_path_info;
       }

       location /govid/static/ {
           alias /var/www/htdocs/govid/;
       }
   }
   ```

3. Install govid:
   ```bash
   sudo make install
   ```

4. Restart Nginx:
   ```bash
   sudo systemctl restart nginx
   ```

### Standalone Mode (Development)

Test locally without a web server:
```bash
make run
```
-> Access at [http://localhost:8080/filename.txt](http://localhost:8080/filename.txt)

Notes are stored in the `./notes` directory (relative to the binary).

## :file_folder: Project Structure

```
govid/
├── govid.go          # Backend (~50 lines)
├── govid_test.go     # Unit tests
├── htdocs/
│   ├── index.html     # Frontend with textarea
│   └── vi-minimal.js  # Ultra-lightweight Vi emulator (80 lines)
├── Makefile          # Build and install scripts
└── README.md         # Documentation
```

## :shield: Security

`govid` implements the following protections:

| Threat | Protection |
|--------|-------------|
| **Path Traversal** (`../etc/passwd`) | Path validation with `strings.Contains(filepath, "..")` |
| **Hidden Files** (`.bashrc`, `.git/`) | Block paths starting with `/.` |
| **Binary Files** | Null byte detection (`0x00`) |
| **Oversized Files** | 1 MB limit per file |
| **Writing Outside `GOVID_DIR`** | Verify absolute path starts with `GOVID_DIR` |

**Additional Recommendations:**
- Use **HTTPS** in production.
- Set directory permissions:
  ```bash
  chown -R www-data:www-data /var/www/notes
  chmod 750 /var/www/notes
  ```
- For basic authentication, use a reverse proxy (Nginx, Apache) or `.htaccess`.

## :computer: Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GOVID_DIR` | Directory for notes storage | `DOCUMENT_ROOT` or `./notes` | `/var/www/notes` |
| `GOVID_URL_PREFIX` | URL prefix to access govid | `SCRIPT_NAME` | `/govid` |
| `DOCUMENT_ROOT` | (Apache) Document root | - | `/var/www/html` |

## :pencil2: Usage

### Edit a File

1. Access the file URL:
   ```
   http://your-server/govid/notes.txt
   ```
2. Press `i` to enter **insert mode**.
3. Edit the text.
4. Press `Esc` to return to **normal mode**.
5. Type `:w` and press `Enter` to **save**.

### Open Another File

Use `:e filename` in normal mode.

### Create a New File

Access a non-existent URL:
```
http://your-server/govid/new-file.txt
```
The editor opens with empty content. Press `i` to insert text, then `:w` to save.

### Create a Directory

Use a path with subdirectories:
```
http://your-server/govid/folder/new-file.txt
```
Parent directories are created automatically.

### Default File

Accessing the root (`/govid/`) opens `index.txt`.

## :test_tube: Tests

Run unit tests:
```bash
make test
```

Tests cover:
- Path validation (path traversal, hidden files)
- File and directory creation
- Size limit (1 MB)
- Binary file detection
- Default file behavior (`index.txt`)

## :lock: Security Audit

Run static security analysis:
```bash
make audit
```

This runs:
- [staticcheck](https://staticcheck.io/) (static analysis)
- [errcheck](https://github.com/kisielk/errcheck) (unhandled error detection)
- [gosec](https://github.com/securego/gosec) (vulnerability detection)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) (known vulnerability detection)

## :page_facing_up: License

This project is licensed under the **BSD 2-Clause License**. See [LICENSE](LICENSE) for details.
