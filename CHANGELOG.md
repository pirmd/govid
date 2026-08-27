# Changelog

Toutes les modifications notables de ce projet seront documentées dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/),
et ce projet adhère à [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2025-XX-XX

### ✨ Nouvelle Version Ultra-Simplifiée
- **Refonte complète** : Remplacement de l'architecture complexe (templates serveur, vi.js) par une version minimaliste.
- **Backend** : Réduit à ~50 lignes de Go (au lieu de ~300).
- **Frontend** : Remplacement de vi.js par un `<textarea>` HTML standard + JS minimal (~20 lignes).
- **Zéro dépendance** : Suppression de toutes les dépendances externes (vi.js, embed.FS, html/template).

### 🚀 Améliorations
- **Mode standalone** : Ajout d'un mode développement avec serveur intégré (`:8080`).
- **Sécurité renforcée** : Validation des chemins et détection des fichiers binaires conservées.
- **Compatibilité universelle** : Fonctionne avec tous les navigateurs (même sans JavaScript pour les fonctionnalités de base).

### 🗑️ Suppressions
- **vi.js** : L'éditeur vi-like a été supprimé au profit d'un `<textarea>` standard.
- **Templates serveur** : Plus besoin de `html/template` ou `embed.FS`.
- **Listage de dossiers** : Fonctionnalité supprimée (navigation manuelle via les URLs).

### 📝 Documentation
- **README.md** : Réécrit pour refléter la nouvelle architecture.
- **Makefile** : Simplifié pour ne plus dépendre de vi.js.
- **Tests** : Nouveaux tests unitaires pour la version simplifiée.

### 🔧 Migration
- **Backward Compatible** : Non (changement complet d'architecture).
- **Nouvelle structure** :
  ```
  govid/
  ├── cgi-bin/
  │   └── govid          # Backend (binaire)
  │   └── govid.go       # Code source
  │   └── govid_test.go  # Tests
  └── htdocs/
      └── index.html     # Frontend
  ```

---

## [1.0.0] - 2024-01-XX

### ✨ Première Version
- Version initiale avec vi.js et templates serveur.
- Fonctionnalités : Édition de fichiers texte avec un éditeur vi-like dans le navigateur.

---

[2.0.0]: https://github.com/pirmd/govid/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/pirmd/govid/releases/tag/v1.0.0
