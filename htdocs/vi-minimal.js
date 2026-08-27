/**
 * MiniVi - Émulateur Vi ultra-minimaliste (80 lignes)
 * Fonctionnalités : i/Esc, h/j/k/l, x, dd, :w, :q, :e
 * Poids : ~2 Ko (vs ~50 Ko pour vi.js)
 * Dépendances : Aucune
 * Licence : MIT
 */

class MiniVi {
  constructor(textarea, statusEl) {
    this.textarea = textarea;
    this.statusEl = statusEl;
    this.mode = 'normal'; // 'normal' ou 'insert'
    this.commandBuffer = '';
    this.setupEvents();
    this.updateStatus();
  }

  // Initialiser les écouteurs d'événements
  setupEvents() {
    this.textarea.addEventListener('keydown', (e) => this.handleKeyDown(e));
    this.textarea.addEventListener('blur', () => this.focus());
  }

  // Gérer les touches
  handleKeyDown(e) {
    // Mode insertion : seul Esc est intercepté
    if (this.mode === 'insert') {
      if (e.key === 'Escape') {
        e.preventDefault();
        this.mode = 'normal';
        this.updateStatus();
      }
      return;
    }

    // Mode normal : intercepter toutes les commandes
    e.preventDefault();

    // Commandes de base
    switch (e.key) {
      // Basculer en mode insertion
      case 'i':
      case 'a':
        this.mode = 'insert';
        this.updateStatus();
        break;

      // Déplacement
      case 'h':
        this.moveCursor(-1);
        break;
      case 'l':
        this.moveCursor(1);
        break;
      case 'j':
        this.moveCursorDown();
        break;
      case 'k':
        this.moveCursorUp();
        break;
      case '0':
        this.moveToLineStart();
        break;
      case '$':
        this.moveToLineEnd();
        break;

      // Suppression
      case 'x':
        this.deleteChar();
        break;
      case 'd':
        if (e.ctrlKey || e.metaKey) break; // Ignorer Ctrl+D, Cmd+D
        this.deleteLine();
        break;

      // Commandes (deux-points)
      case ':':
        this.commandBuffer = ':';
        this.updateStatus();
        break;

      // Annuler la commande
      case 'Escape':
        this.commandBuffer = '';
        this.updateStatus();
        break;

      // Retour chariot (exécuter la commande)
      case 'Enter':
        if (this.commandBuffer) {
          this.executeCommand(this.commandBuffer.substring(1)); // Enlever le ':'
          this.commandBuffer = '';
          this.updateStatus();
        }
        break;

      // Ajouter au buffer de commande
      default:
        if (this.commandBuffer) {
          this.commandBuffer += e.key;
          this.updateStatus();
        }
    }
  }

  // Déplacer le curseur horizontalement
  moveCursor(offset) {
    const pos = this.textarea.selectionStart + offset;
    const max = this.textarea.value.length;
    this.textarea.selectionStart = Math.max(0, Math.min(pos, max));
    this.textarea.selectionEnd = this.textarea.selectionStart;
    this.textarea.focus();
  }

  // Déplacer le curseur vers le début de la ligne
  moveToLineStart() {
    const lines = this.textarea.value.split('\n');
    const currentPos = this.textarea.selectionStart;
    let line = 0;
    let posInLine = currentPos;

    for (let i = 0; i < lines.length; i++) {
      const lineLength = lines[i].length + (i < lines.length - 1 ? 1 : 0);
      if (posInLine <= lineLength) {
        line = i;
        break;
      }
      posInLine -= lineLength;
    }

    let newPos = 0;
    for (let i = 0; i < line; i++) {
      newPos += lines[i].length + 1;
    }

    this.textarea.selectionStart = newPos;
    this.textarea.selectionEnd = newPos;
    this.textarea.focus();
  }

  // Déplacer le curseur vers la fin de la ligne
  moveToLineEnd() {
    const lines = this.textarea.value.split('\n');
    const currentPos = this.textarea.selectionStart;
    let line = 0;
    let posInLine = currentPos;

    for (let i = 0; i < lines.length; i++) {
      const lineLength = lines[i].length + (i < lines.length - 1 ? 1 : 0);
      if (posInLine <= lineLength) {
        line = i;
        break;
      }
      posInLine -= lineLength;
    }

    let newPos = 0;
    for (let i = 0; i < line; i++) {
      newPos += lines[i].length + 1;
    }
    newPos += lines[line].length;

    this.textarea.selectionStart = newPos;
    this.textarea.selectionEnd = newPos;
    this.textarea.focus();
  }

  // Déplacer le curseur vers le bas
  moveCursorDown() {
    const lines = this.textarea.value.split('\n');
    const currentPos = this.textarea.selectionStart;
    let line = 0;
    let posInLine = currentPos;

    // Trouver la ligne actuelle
    for (let i = 0; i < lines.length; i++) {
      const lineLength = lines[i].length + (i < lines.length - 1 ? 1 : 0);
      if (posInLine <= lineLength) {
        line = i;
        break;
      }
      posInLine -= lineLength;
    }

    // Passer à la ligne suivante (si possible)
    if (line < lines.length - 1) {
      const nextLineLength = lines[line + 1].length;
      const newPosInLine = Math.min(posInLine, nextLineLength);
      let newPos = 0;

      // Calculer la position absolue
      for (let i = 0; i <= line; i++) {
        newPos += lines[i].length + 1; // +1 pour '\n'
      }
      newPos += newPosInLine;

      this.textarea.selectionStart = newPos;
      this.textarea.selectionEnd = newPos;
    }
    this.textarea.focus();
  }

  // Déplacer le curseur vers le haut
  moveCursorUp() {
    const lines = this.textarea.value.split('\n');
    const currentPos = this.textarea.selectionStart;
    let line = 0;
    let posInLine = currentPos;

    // Trouver la ligne actuelle
    for (let i = 0; i < lines.length; i++) {
      const lineLength = lines[i].length + (i < lines.length - 1 ? 1 : 0);
      if (posInLine <= lineLength) {
        line = i;
        break;
      }
      posInLine -= lineLength;
    }

    // Remonter à la ligne précédente (si possible)
    if (line > 0) {
      const prevLineLength = lines[line - 1].length;
      const newPosInLine = Math.min(posInLine, prevLineLength);
      let newPos = 0;

      // Calculer la position absolue
      for (let i = 0; i < line - 1; i++) {
        newPos += lines[i].length + 1;
      }
      newPos += newPosInLine;

      this.textarea.selectionStart = newPos;
      this.textarea.selectionEnd = newPos;
    }
    this.textarea.focus();
  }

  // Supprimer le caractère sous le curseur
  deleteChar() {
    const pos = this.textarea.selectionStart;
    if (pos < this.textarea.value.length) {
      this.textarea.value =
        this.textarea.value.substring(0, pos) +
        this.textarea.value.substring(pos + 1);
      this.textarea.selectionStart = pos;
      this.textarea.selectionEnd = pos;
    }
  }

  // Supprimer la ligne actuelle
  deleteLine() {
    const lines = this.textarea.value.split('\n');
    const currentPos = this.textarea.selectionStart;
    let lineStart = 0;
    let lineEnd = 0;
    let lineIndex = 0;

    // Trouver les indices de la ligne actuelle
    for (let i = 0; i < lines.length; i++) {
      const lineLength = lines[i].length + (i < lines.length - 1 ? 1 : 0);
      if (currentPos <= lineStart + lineLength) {
        lineIndex = i;
        lineEnd = lineStart + lineLength;
        break;
      }
      lineStart += lineLength;
    }

    // Supprimer la ligne
    const newValue = lines.filter((_, i) => i !== lineIndex).join('\n');
    this.textarea.value = newValue;

    // Positionner le curseur au début de la ligne suivante (ou précédente)
    this.textarea.selectionStart = Math.min(lineStart, newValue.length);
    this.textarea.selectionEnd = this.textarea.selectionStart;
  }

  // Exécuter une commande (ex: w, q, e filename)
  executeCommand(cmd) {
    cmd = cmd.trim();
    if (!cmd) return;

    const parts = cmd.split(/\s+/);
    const action = parts[0];
    const args = parts.slice(1);

    switch (action) {
      case 'w':
        this.saveFile();
        break;
      case 'q':
        window.close();
        break;
      case 'e':
        if (args[0]) {
          const currentPath = window.location.pathname;
          const newPath = currentPath.replace(/([^/]+)$/, args[0]);
          window.location.pathname = newPath;
        }
        break;
      default:
        this.statusEl.textContent = `Commande inconnue: :${cmd}`;
    }
  }

  // Sauvegarder le fichier
  saveFile() {
    this.statusEl.textContent = "Sauvegarde...";
    const form = this.textarea.closest('form');
    if (form) {
      form.requestSubmit(); // Soumet le formulaire parent
    }
  }

  // Mettre à jour la barre de statut
  updateStatus() {
    this.statusEl.className = 'status ' + this.mode;
    let status = `Mode: ${this.mode === 'insert' ? 'INSERT' : 'NORMAL'}`;
    if (this.commandBuffer) {
      status += ` | Commande: ${this.commandBuffer}`;
    }
    this.statusEl.textContent = status;
  }

  // Mettre le focus sur le textarea
  focus() {
    this.textarea.focus();
  }
}

// Initialiser MiniVi quand le DOM est chargé
document.addEventListener('DOMContentLoaded', () => {
  const textarea = document.getElementById('content');
  const statusEl = document.getElementById('status');
  if (textarea && statusEl) {
    new MiniVi(textarea, statusEl);
  }
});
