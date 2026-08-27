# GOVID - Éditeur de Notes Minimaliste avec Mode Vi

[![Go Reference](https://pkg.go.dev/badge/github.com/pirmd/govid.svg)](https://pkg.go.dev/github.com/pirmd/govid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirmd/govid)](https://goreportcard.com/report/github.com/pirmd/govid)

`govid` est une application **ultra-minimaliste** pour éditer des fichiers texte directement depuis un navigateur, avec un **émulateur Vi léger** (80 lignes de JS) pour les utilisateurs habitués à vi/vim.

Cette version allie **simplicité extrême** (backend en 50 lignes de Go) et **fonctionnalités Vi basiques** pour une expérience familière.

## 🎯 Fonctionnalités

### ✅ Fonctionnalités de Base
- Édition de fichiers texte via un navigateur
- Création automatique de dossiers
- Validation de sécurité (path traversal, fichiers cachés, fichiers binaires)
- Limite de taille (1 Mo par fichier)
- Mode CGI (Apache/Nginx) et mode standalone (développement sur `:8080`)

### ✅ Fonctionnalités Vi (MiniVi)
| Commande | Description |
|----------|-------------|
| `i` | Entrer en mode insertion |
| `a` | Entrer en mode insertion (après le curseur) |
| `Esc` | Retour au mode normal |
| `h` | Déplacer le curseur à gauche |
| `j` | Déplacer le curseur vers le bas |
| `k` | Déplacer le curseur vers le haut |
| `l` | Déplacer le curseur à droite |
| `0` | Aller au début de la ligne |
| `$` | Aller à la fin de la ligne |
| `x` | Supprimer le caractère sous le curseur |
| `dd` | Supprimer la ligne actuelle |
| `:w` | Sauvegarder le fichier |
| `:q` | Quitter (fermer l'onglet) |
| `:e fichier` | Ouvrir un autre fichier |

## 🚀 Installation

### Dépendances
- [Go 1.17+](https://golang.org/dl/) (pour le build)
- Un serveur web (Apache, Nginx, Caddy) **ou** aucun serveur pour le mode standalone

### Build
```bash
make
```
Cela génère le binaire `govid`.

### Déploiement avec Apache

1. Installez Apache et activez le module CGI :
   ```bash
   sudo apt install apache2
   sudo a2enmod cgi
   ```

2. Configurez Apache pour utiliser `govid` :
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

3. Installez govid :
   ```bash
   sudo make install
   ```

4. Redémarrez Apache :
   ```bash
   sudo systemctl restart apache2
   ```

5. Accédez à [http://votre-serveur/govid/nom-du-fichier.txt](http://votre-serveur/govid/nom-du-fichier.txt)

### Déploiement avec Nginx

1. Installez Nginx et fcgiwrap :
   ```bash
   sudo apt install nginx fcgiwrap
   sudo systemctl start fcgiwrap.socket
   sudo systemctl enable fcgiwrap.socket
   ```

2. Configurez Nginx :
   ```nginx
   # /etc/nginx/sites-available/govid
   server {
       listen 80;
       server_name votre-serveur.com;

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

3. Installez govid :
   ```bash
   sudo make install
   ```

4. Redémarrez Nginx :
   ```bash
   sudo systemctl restart nginx
   ```

### Mode Standalone (Développement)

Pour tester localement sans serveur web :
```bash
make run
```
→ Accédez à [http://localhost:8080/nom-du-fichier.txt](http://localhost:8080/nom-du-fichier.txt)

Les notes seront stockées dans le dossier `./notes` (relatif au binaire).

## 📁 Structure du Projet

```
govid/
├── govid.go          # Backend (50 lignes)
├── govid_test.go     # Tests unitaires
├── htdocs/
│   ├── index.html     # Frontend avec textarea
│   └── vi-minimal.js  # Émulateur Vi ultra-léger (80 lignes)
├── Makefile           # Scripts de build et installation
└── README.md          # Documentation
```

## 🛡️ Sécurité

`govid` implémente les protections suivantes :

| Menace | Protection |
|--------|-------------|
| **Path Traversal** (`../etc/passwd`) | Validation des chemins avec `strings.Contains(filepath, "..")` |
| **Accès aux fichiers cachés** (`.bashrc`, `.git/`) | Blocage des chemins commençant par `/.` |
| **Fichiers binaires** | Détection des bytes nuls (`0x00`) |
| **Taille excessive** | Limite à **1 Mo** par fichier |
| **Écriture en dehors de `GOVID_DIR`** | Vérification que le chemin absolu commence par `GOVID_DIR` |

**Recommandations supplémentaires :**
- Utilisez **HTTPS** en production.
- Configurez les **permissions** du dossier de notes :
  ```bash
  chown -R www-data:www-data /var/www/notes
  chmod 750 /var/www/notes
  ```
- Pour une **authentification basique**, utilisez un reverse proxy (Nginx, Apache) ou un `.htaccess`.

## 📂 Variables d'Environnement

| Variable | Description | Valeur par défaut | Exemple |
|----------|-------------|------------------|---------|
| `GOVID_DIR` | Dossier où sont stockées les notes | `DOCUMENT_ROOT` ou `./notes` | `/var/www/notes` |
| `GOVID_URL_PREFIX` | Préfixe URL pour accéder à govid | `SCRIPT_NAME` | `/govid` |
| `DOCUMENT_ROOT` | (Apache) Racine des documents | - | `/var/www/html` |

## 📖 Utilisation

### Éditer un fichier

1. Accédez à l'URL du fichier :
   ```
   http://votre-serveur/govid/notes.txt
   ```
2. Appuyez sur `i` pour entrer en **mode insertion**.
3. Modifiez le texte.
4. Appuyez sur `Esc` pour revenir en **mode normal**.
5. Tapez `:w` et appuyez sur `Entrée` pour **sauvegarder**.

### Commandes Vi Disponibles

| Commande | Description |
|----------|-------------|
| `i` | Entrer en mode insertion (avant le curseur) |
| `a` | Entrer en mode insertion (après le curseur) |
| `Esc` | Retour au mode normal |
| `h` | Déplacer le curseur à gauche |
| `j` | Déplacer le curseur vers le bas |
| `k` | Déplacer le curseur vers le haut |
| `l` | Déplacer le curseur à droite |
| `0` | Aller au début de la ligne |
| `$` | Aller à la fin de la ligne |
| `x` | Supprimer le caractère sous le curseur |
| `dd` | Supprimer la ligne actuelle |
| `:w` | Sauvegarder le fichier |
| `:q` | Quitter (fermer l'onglet) |
| `:e nom-fichier` | Ouvrir un autre fichier |

### Créer un nouveau fichier

- Accédez à une URL inexistante :
  ```
  http://votre-serveur/govid/nouveau-fichier.txt
  ```
- L'éditeur s'ouvrira avec un contenu vide.
- Appuyez sur `i` pour insérer du texte, puis `:w` pour sauvegarder.

### Créer un dossier

- Utilisez un chemin avec des sous-dossiers :
  ```
  http://votre-serveur/govid/dossier/nouveau-fichier.txt
  ```
- Les dossiers parents seront créés automatiquement.

### Fichier par défaut

- Si vous accédez à la racine (`/govid/`), le fichier `index.txt` sera ouvert.

## 🧪 Tests

Pour exécuter les tests unitaires :
```bash
make test
```

Les tests couvrent :
- Validation des chemins (path traversal, fichiers cachés)
- Création de fichiers et dossiers
- Limite de taille (1 Mo)
- Détection des fichiers binaires
- Comportement par défaut (fichier `index.txt`)

## 🔧 Audit de Sécurité

Pour exécuter une analyse de sécurité statique :
```bash
make audit
```

Cela exécute :
- [staticcheck](https://staticcheck.io/) (analyse statique)
- [errcheck](https://github.com/kisielk/errcheck) (détection des erreurs non gérées)
- [gosec](https://github.com/securego/gosec) (détection des vulnérabilités)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) (détection des vulnérabilités connues)

## 📜 Changelog

Consultez [CHANGELOG.md](CHANGELOG.md) pour l'historique des versions.

## 🤝 Contribution

Les contributions sont les bienvenues ! Ouvrez une issue ou une pull request sur [GitHub](https://github.com/pirmd/govid).

## 📄 Licence

Ce projet est sous licence MIT. Consultez [LICENSE](LICENSE) pour plus de détails.
