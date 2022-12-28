const CMD_MODE = 'COMMAND';
const INS_MODE = 'INSERT';
const EX_MODE = 'EX';

const DOWN = 1;
const UP = -1;
const LEFT = -1;
const RIGHT = 1;

function VI(textarea, status) {
    this.setMode = function(mode) {
        if (mode === EX_MODE) {
            this.ex.disabled = false;
            this.ex.focus();
        } else {
            this.ex.disabled = true;
            this.ed.textarea.focus();
            this.ed.snapshot();
        };

        if (this.mode === INS_MODE) this.ex.value = "";
        if (mode === INS_MODE) this.ex.value = "-- INSERT --";

        this.mode = mode;
    };

    this.onkeydown = function(event) {
        let k = event.key;

        if (event.ctrlKey) {
            k = "Ctrl-" + k;
        };

        if (!this.handleKey(k)) {
            event.preventDefault();
        };
    };

    this.handleKey = function(k) {
        switch (true) {
            case (k === "Escape"):
                this.setMode(CMD_MODE);
                this.ctree.reset();
                return false;
            case (this.mode == INS_MODE):
                if (k === "Enter") {
                    this.ed.insert("\n");
                    this.ed.moveV(DOWN);
                    this.ed.indent();
                    return false;
                };
                return true;
            case (this.mode == EX_MODE):
                if (k === "Enter") {
                    this.execute(this.ex.value);
                    this.setMode(CMD_MODE);
                    return false;
                };
                return true;
            default: // CMD_MODE
                if ((k === ":") || (k === "/") || (k === "?")) {
                    this.setMode(EX_MODE);
                    return true;
                };

                this.ctree.handleKey(k);
                return false;
        };
    };

    this.execute = function(cmdline) {
        const args = cmdline.split(" ");
        switch (true) {
            case (args[0] === ":w"):
                try {
                    this.save(...args.slice(1));
                    this.ex.value = "";
                } catch(err) {
                    this.err(err);
                };
                return;

            case (args[0] === ":q"):
                try {
                    this.quit();
                    this.ex.value = "";
                } catch(err) {
                    this.err(err);
                };
                return;

            case (args[0] === ":wq"):
                try {
                    this.save();
                    this.ex.value = "";
                } catch(err) {
                    this.err(err);
                    return;
                };
                try {
                    this.quit();
                    this.ex.value = "";
                } catch(err) {
                    this.err(err);
                };
                return;

            case (cmdline.startsWith("/")):
                let pos = this.ed.search(cmdline.slice(1));
                if (pos === -1) {
                    this.err( (this.lastsearch === "") ? "No previous search pattern" : "Pattern not found" );
                } else {
                    this.ed.moveTo(pos);
                    this.ex.value = "";
                };
                return;

            case (cmdline.startsWith("?")):
                pos = this.ed.searchBackwards(cmdline.slice(1));
                if (pos === -1) {
                    this.err( (this.lastsearch === "") ? "No previous search pattern" : "Pattern not found");
                } else {
                    this.ed.moveTo(pos);
                    this.ex.value = "";
                };
                return;

            default:
                this.err("Not an editor command");
        };
    };

    this.copyToClipboard = function(start, end) {
        start = (typeof(start) === "undefined") ? this.ed.cursor() : start;
        end = (typeof(end) === "undefined") ? start + 1 : end;

        this.clipboard = this.ed.text().slice(start, end);
    };

    this.pasteFromClipboard = function(start, end) {
        this.ed.insert(this.clipboard, start, end);
    };

    this.err = function(msg) {
        this.ex.value = msg;
        setTimeout(() => this.ex.value="", 1000);
    };

    this.maxhistory = 64;
    this.shiftwidth = 4;
    this.save = () => { throw "Save is not implemented"; };
    this.quit = () => { throw "Quit is not implemented"; };

    this.mode = CMD_MODE;
    this.ctree = new CommandTree(this, cmdModeCtree);
    this.clipboard = "";

    textarea.onkeydown = this.onkeydown.bind(this);
    this.ed = new Textarea(textarea, this.maxhistory);
    this.ed.snapshot();

    status.onkeydown = this.onkeydown.bind(this);
    this.ex = status;
};


function Textarea(textarea, maxHistory) {
    this.textarea = textarea;
    this.history = new Array();
    this.historyRedo = new Array();
    this.maxHistory = (typeof(maxHistory) === "undefined") ? 64 : maxHistory;
    this.lastsearch = "";

    this.text = function() {
        return this.textarea.value;
    };
    this.EOF = function() {
        return this.text().length - 1;
    };

    this.cursor = function() {
        return this.textarea.selectionStart;
    };

    this.moveTo = function(pos) {
        this.textarea.setSelectionRange(pos, pos);
    };

    this.moveH = function(di) {
        let pos = this.cursor() + di;

        let curL = this.line();
        pos = (pos < curL.start) ? curL.start : pos;
        pos = (pos > curL.end) ? curL.end : pos;

        this.moveTo(pos);
    };

    this.moveV = function(dj) {
        let curL = this.line();
        if (typeof(curL) === "undefined") return;

        let targetL = this.lineV(dj, curL);
        if (typeof(targetL) === "undefined") return;

        let pos = targetL.start + (this.cursor() - curL.start);
        pos = (pos > targetL.end) ? targetL.end : pos;

        this.moveTo(pos);
    };

    this.line = function(pos) {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;

        if (pos < 0 || pos > this.EOF()) {
            return undefined;
        };

        let start = (this.text().charAt(pos) === "\n") ?
            (pos > 0) ? this.text().lastIndexOf("\n", pos-1) + 1 : 0 :
            this.text().lastIndexOf("\n", pos) + 1;

        let end = this.text().indexOf("\n", pos);

        return {
            'start': start,
            'end': (end === -1) ? this.EOF() : end,
        };
    };

    this.lineV = function(dj, fromL) {
        let nextL = (typeof(fromL) === "undefined") ? this.line() : fromL;
        if (typeof(nextL) === "undefined") return undefined;

        for (let j = dj; j !== 0; (j > 0) ? j-- : j++) {
            nextL = (j > 0) ? this.line(nextL.end+1) : this.line(nextL.start-1);
            if (typeof(nextL) === "undefined") break;
        };

        return nextL;
    };

    this.word = function (pos) {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;
        let c = this.text().charAt(pos);
        if (/(\p{Z}|\n)/u.test(c)) return undefined;
        if (/\p{P}/u.test(c)) return {start: pos, end: pos};

        let end = pos;
        for (end = pos+1; end <= this.EOF(); end++) {
            let c = this.text().charAt(end);
            if (/(\n|\p{Z}|\p{P})/u.test(c)) {
                end--;
                break;
            };
        };

        let start = pos;
        for (start = pos-1; start > 0; start--) {
            let c = this.text().charAt(start);
            if (/(\n|\p{Z}|\p{P})/u.test(c)) {
                start++;
                break;
            };
        };

        return {start: start, end: end};
    };

    this.nextWord = function(pos) {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;
        let curW = this.word(pos);
        let curP = (typeof(curW) === "undefined") ? pos : curW.end + 1;

        let nextW = undefined
        for (let p = curP; p <= this.EOF(); p++) {
            nextW = this.word(p);
            if (typeof(nextW) !== "undefined") break;
        };

        return nextW;
    };
    
    this.insert = function(txt, start, end) {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        end = (typeof(end) === "undefined") ? start : end;

        this.snapshot();
        this.textarea.setRangeText(txt, start, end);
        this.textarea.dispatchEvent(new Event("input"));
    };

    this.delete = function(start, end) {
        this.insert("", start, end);
    };

    this.search = function(value, start) {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        if (value === "") value = this.lastsearch;
        if (value === "") return -1;

        let txt = this.textarea.value.slice(start);
        let pos = txt.indexof(value);
        return (pos === -1) ? -1 : pos + start;
    };

    this.searchBackwards = function(value, start) {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        if (value === "") value = this.lastsearch;
        if (value === "") return -1;

        let txt = this.textarea.value.slice(0, start);
        let pos = txt.lastindexof(value);
        return ( pos === -1) ? -1 : pos;
    };

    this.indent = function() {
        let curL = this.line();
        if (curL.start === 0) return;

        for (var i = curL.start; i < curL.end; i++) {
            let c = this.text().charAt(i);
            if (!/(\p{Z})/u.test(c)) break;
        };
        let curSP = (i - curL.start);

        let prevL = this.line(curL.start - 1);
        for (var j = prevL.start; j < prevL.end; j++) {
            let c = this.text().charAt(j);
            if (!/(\p{Z})/u.test(c)) break;
        };
        let prevSP = (j - prevL.start);

        if (curSP < prevSP) {
            let sp = prevSP - curSP;
            this.insert(" ".repeat(sp), curL.start);
            this.moveTo(i + sp);
        } else if (curSP > prevSP) {
            let sp = curSP - prevSP
            this.delete(curL.start, curL.start + sp);
            this.moveTo(i - sp);
        };
    };

    this.snapshot = function() {
        if (this.history.length > 0 && (this.text() === this.history[this.history.length - 1])) return;

        this.history.push({
            text: this.text(),
            pos:  this.cursor(),
        });

        if (this.history.length > this.maxHistory) {
            this.history.shift();
        };
    };

    this.undo = function() {
        if (this.history.length > 0) {
            this.historyRedo.push({
                text: this.text(),
                pos:  this.cursor(),
            });
            if (this.historyRedo.length > this.maxHistory) {
                this.historyRedo.shift();
            };

            let u = this.history.pop();
            this.textarea.value = u.text;
            this.moveTo(u.pos);
        };
    };

    this.redo = function() {
        if (this.historyRedo.length > 0) {
            this.snapshot();

            let u = this.historyRedo.pop();
            this.textarea.value = u.text;
            this.moveTo(u.pos);
        };
    };
};

function CommandTree(vi, ctree) {
    this.vi = vi;
    this.tree = ctree;
    this.curNode = this.tree;
    this.counter = "";
    this.timeoutId = null;
    this.lastCmd = null;

    this.reset = function() {
        clearTimeout(this.timeoutId);
        this.curNode = this.tree;
        this.counter = "";
    };

    this.handleKey = function(k) {
        clearTimeout(this.timeoutId);
        this.timeoutId = setTimeout(() => {this.reset()}, 800);

        if ((k >= '0') && (k <= '9')) {
            // entering a number/count reset the current sequence of keys
            this.curNode = this.tree;

            // '0' as command for reaching start of line
            if (!((k === '0') && (this.counter === ''))) {
                this.counter += k;
                return;
            };
        };

        this.curNode = this.curNode.getNextNode(k);
        if (typeof(this.curNode) === "undefined") {
            this.reset();
            return;
        };

        if (typeof(this.curNode.action) !== "undefined") {
            let c = parseInt(this.counter, 10);
            let cmd = this.curNode.action(this.vi, (isNaN(c)) ? 0 : c);
            this.lastCmd = (cmd) ? cmd : {key: k, action: this.curNode.action};
            this.reset();
        };
    };
};

function KeyNode(action) {
    this.action = action;
    this.nextNodes = new Map();

    this.getNextNode = function(k) {
        return this.nextNodes.get(k);
    };

    this.addNextNode = function(k, n) {
        this.nextNodes.set(k, n);
        return this;
    };
};

var cmdModeCtree = new KeyNode()
    .addNextNode("h", new KeyNode((vi, c) => vi.ed.moveH((c == 0) ? LEFT : LEFT * c)))
    .addNextNode("j", new KeyNode((vi, c) => vi.ed.moveV((c == 0) ? DOWN : DOWN * c)))
    .addNextNode("k", new KeyNode((vi, c) => vi.ed.moveV( (c == 0) ? UP : UP * c)))
    .addNextNode("l", new KeyNode((vi, c) => vi.ed.moveH( (c == 0) ? RIGHT : RIGHT * c)))
    .addNextNode("0", new KeyNode((vi, c) => vi.ed.moveTo(vi.ed.line().start)))
    .addNextNode("$", new KeyNode((vi, c) => {
        if ( c > 1 ) { vi.ed.moveV(DOWN * (c-1)) };
        vi.ed.moveTo(vi.ed.line().end);
    }))
    .addNextNode("G", new KeyNode((vi, c) => {
        if ( c === 0 ) {
            vi.ed.moveTo(vi.ed.line(vi.ed.EOF()).start);
        } else {
            let l = vi.ed.lineV(c-1, vi.ed.line(0));
            if (typeof(l) !== "undefined") { vi.ed.moveTo(l.start) };
        };
    }))
    .addNextNode("w", new KeyNode((vi, c) => {
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            nextW = vi.ed.nextWord();
            if (typeof(nextW) !== "undefined") {
                vi.ed.moveTo(nextW.start);
            };
        };
    }))
    .addNextNode("i", new KeyNode((vi, c) => vi.setMode(INS_MODE)))
    .addNextNode("a", new KeyNode((vi, c) => {
        vi.ed.moveH(RIGHT);
        vi.setMode(INS_MODE);
    }))
    .addNextNode("A", new KeyNode((vi, c) => {
        vi.ed.moveTo(vi.ed.line().end);
        vi.setMode(INS_MODE);
    }))
    .addNextNode(">", new KeyNode((vi, c) => {
        let pos = vi.ed.cursor();
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            let curL = vi.ed.line(pos);
            vi.ed.insert(" ".repeat(vi.shiftwidth), curL.start);

            if (curL.end === vi.ed.EOF()) break;
            pos = curL.end + 1;
        };
        vi.ed.moveTo(vi.ed.cursor()+vi.shiftwidth);
    }))
    .addNextNode("<", new KeyNode((vi, c) => {
        let pos = vi.ed.cursor();
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            let curL = vi.ed.line(pos);
            for (var ii = curL.start; ii < curL.end; ii++) {
                let c = vi.ed.text().charAt(ii);
                if (!/(\p{Z})/u.test(c)) break;
            };
            let sp = (ii - curL.start) - vi.shiftwidth;
            if (sp > 0) vi.ed.delete(curL.start, curL.start + sp);

            if (curL.end === vi.ed.EOF()) break;
            pos = curL.end + 1;
        };
    }))
    .addNextNode("J", new KeyNode((vi, c) => {
        let curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        if (curL.end === vi.ed.EOF()) return;

        vi.ed.delete(curL.end);
        vi.ed.moveV(DOWN);
        vi.setMode(INS_MODE);
    }))
    .addNextNode("o", new KeyNode((vi, c) => {
        let curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;

        if (vi.ed.text().charAt(curL.end) !== "\n") vi.ed.insert("\n", curL.end + 1);
        vi.ed.insert("\n", curL.end + 1);
        vi.ed.moveV(DOWN);
        vi.setMode(INS_MODE);
    }))
    .addNextNode("O", new KeyNode((vi, c) => {
        let curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;

        vi.ed.insert("\n", curL.start);
        vi.ed.moveTo(curL.start - 1);
        vi.setMode(INS_MODE);
    }))
    .addNextNode("x", new KeyNode((vi, c) => {
        let start = vi.ed.cursor();
        let end = ( c > 0 ) ? start + c : start + 1; 

        curL = vi.ed.line();
        if ( end > curL.end) { end = curL.end; };
        
        vi.copyToClipboard(start, end);
        vi.ed.delete(start, end);
    }))
    .addNextNode("D", new KeyNode((vi, c) => {
        let curL = vi.ed.line();
        vi.copyToClipboard(pos, curL.end);
        vi.ed.delete(pos, curL.end);
    }))
    .addNextNode("d", new KeyNode()
        .addNextNode("d", new KeyNode((vi, c) => {
            let curL = vi.ed.line();
            let endL = (c > 1) ? vi.ed.lineV(c-1) : curL;

            vi.copyToClipboard(curL.start, endL.end+1);
            if ( endL.end === vi.ed.EOF() ) {
                vi.ed.delete(curL.start - 1, endL.end);
            } else {
                vi.ed.delete(curL.start, endL.end+1);
            };
            vi.ed.moveTo(curL.start);
        }))
        .addNextNode("w", new KeyNode((vi, c) => {
            let start = vi.ed.cursor();
            let end = -1;
            for (let i = (c === 0)? 1 : c; i > 0; i--) {
                let nextW = vi.ed.nextWord();
                if (typeod(nextW) === "undefined") break;
                end = nextW.start;
            };
            if (end === -1) return;
            vi.copyToClipboard(start, end);
            vi.ed.delete(start, end);
        }))
    )
    .addNextNode("y", new KeyNode()
        .addNextNode("y", new KeyNode((vi, c) => {
            let curL = vi.ed.line();
            let endL = (c > 1) ? vi.ed.lineV(c-1) : curL;
            vi.copyToClipboard(curL.start, endL.end+1);
        }))
        .addNextNode("w", new KeyNode((vi, c) => {
            let start = vi.ed.cursor();
            let end = -1;
            for (let i = (c === 0)? 1 : c; i > 0; i--) {
                let nextW = vi.ed.nextWord();
                if (typeod(nextW) === "undefined") break;
                end = nextW.start;
            };
            if (end === -1) return;
            vi.copyToClipboard(start, end);
        }))
    )
    .addNextNode("p", new KeyNode((vi, c) => {
        if (vi.clipboard[vi.clipboard.length - 1] === "\n") {
            let start = vi.ed.line().end + 1;
            if (start >= vi.ed.EOF()) {
                vi.ed.insert("\n", start);
            };
            vi.pasteFromClipboard(start);
        } else {
            vi.pasteFromClipboard(vi.ed.cursor() + 1);
        };
    }))
    .addNextNode("P", new KeyNode((vi, c) => {
        if (vi.clipboard[vi.clipboard.length - 1] === "\n") {
            let start = vi.ed.line().start;
            vi.pasteFromClipboard(start);
        } else {
            let start = vi.ed.cursor();
            vi.pasteFromClipboard((start === 0) ? start : start - 1);
        };
    }))
    .addNextNode("u", new KeyNode((vi, c) => {
        if (vi.ctree.lastCmd && vi.ctree.lastCmd.key === "u_undo") {
            vi.ed.redo();
            return {key: "u_redo", action: (vi, c) => {vi.ed.undo()}};
        };

        vi.ed.undo();
        return {key: "u_undo", action: (vi, c) => {vi.ed.undo()}};
    }))
    .addNextNode(".", new KeyNode((vi, c) => {
        if (vi.ctree.lastCmd) vi.ctree.lastCmd.action(vi, c);
        return vi.ctree.lastCmd;
    }));
