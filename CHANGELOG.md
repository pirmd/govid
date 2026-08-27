# Changelog

Toutes les modifications notables de ce projet seront documentées dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/),
et ce projet adhère à [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2025-XX-XX

### ✨ Nouvelle Version Ultra-Simplifiée avec MiniVi

- **Refonte complète** : Remplacement de l'architecture complexe (vi.js, templates serveur) par une version minimaliste.
- **Backend** : Réduit à ~50 lignes de Go (au lieu de ~300).
- **Frontend** : Remplacement de vi.js (~500+ lignes) par un émulateur Vi ultra-léger (**~80 lignes**).
- **Zéro dépendance** : Suppression de toutes les dépendances externes (vi.js, embed.FS, html/template).

### 🚀 Améliorations

- **Mode standalone** : Ajout d'un mode développement avec serveur intégré (`:8080`).
- **Sécurité renforcée** : Validation des chemins et détection des fichiers binaires conservées.
- **Compatibilité universelle** : Fonctionne avec tous les navigateurs.
- **Expérience Vi** : Commandes Vi basiques implémentées (i, Esc, h/j/k/l, x, dd, :w, :q, :e).

### 🗑️ Suppressions

- **vi.js** : L'éditeur vi-like lourd a été supprimé.
- **Templates serveur** : Plus besoin de `html/template` ou `embed.FS`.
- **Listage de dossiers** : Fonctionnalité supprimée (navigation manuelle via les URLs).

### 📝 Documentation

- **README.md** : Réécrit pour refléter la nouvelle architecture et documenter les commandes Vi.
- **Makefile** : Simplifié pour ne plus dépendre de vi.js.
- **Tests** : Nouveaux tests unitaires pour la version simplifiée.

### 🔧 Migration

- **Backward Compatible** : Non (changement complet d'architecture).
- **Nouvelle structure** :
  ```
  govid/
  ├── govid.go          # Backend (50 lignes)
  ├── govid_test.go     # Tests unitaires
  ├── htdocs/
  │   ├── index.html     # Frontend avec textarea
  │   └── vi-minimal.js  # Émulateur Vi (80 lignes)
  └── Makefile          # Build et installation
  ```

---

## [1.0.0] - 2024-01-XX

### ✨ Première Version

- Version initiale avec vi.js et templates serveur.
- Fonctionnalités : Édition de fichiers texte avec un éditeur vi-like dans le navigateur.

---

[2.0.0]: https://github.com/pirmd/govid/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/pirmd/govid/releases/tag/v1.0.0
