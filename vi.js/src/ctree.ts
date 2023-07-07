type Action = {key: string, action: ActionFn}
type ActionFn = (vi: VI, c: number) => Action | void;

class KeyNode {
    action: ActionFn | undefined;
    next: Map<string, KeyNode> = new Map();

    constructor(action?: ActionFn) {
        this.action = action;
    }

    get(k: string): KeyNode | undefined {
        return this.next.get(k);
    }

    add(k: string, node: KeyNode): KeyNode {
        this.next.set(k, node);
        return this;
    }
}

class CommandTree {
    vi: VI;
    root: KeyNode;
    curNode: KeyNode;
    counter = "";
    timeoutId: number|undefined = undefined;
    lastCmd: Action|undefined = undefined;

    constructor(vi: VI) {
        this.vi = vi;
        this.root = commands;
        this.curNode = this.root;
    }

    reset() {
        this.curNode = this.root;
        this.counter = "";
        clearTimeout(this.timeoutId);
        this.timeoutId = undefined;
    }

    handleKey(k: string) {
        clearTimeout(this.timeoutId);
        this.timeoutId = setTimeout(() => {this.reset()}, 800);

        if ((k >= '0') && (k <= '9')) {
            // entering a number/count reset the current sequence of keys
            this.curNode = this.root;

            // '0' as command for reaching start of line
            if (!((k === '0') && (this.counter === ''))) {
                this.counter += k;
                return;
            }
        }

        const node = this.curNode.get(k);
        if (typeof(node) === "undefined") {
            this.reset();
            return;
        }
        this.curNode = node;

        if (typeof(this.curNode.action) !== "undefined") {
            const c = parseInt(this.counter, 10);
            const cmd = this.curNode.action(this.vi, (isNaN(c)) ? 0 : c);
            this.lastCmd = (cmd) ? cmd : {key: k, action: this.curNode.action};
            this.reset();
        }
    }
}

const commands = new KeyNode()
    .add("h", new KeyNode((vi: VI, c: number) => {
        vi.ed.moveH((c == 0) ? -1 : -c)
    }))
    .add("j", new KeyNode((vi: VI, c: number) => {
        vi.ed.moveV((c == 0) ? 1 : c)
    }))
    .add("k", new KeyNode((vi: VI, c: number) => {
        vi.ed.moveV( (c == 0) ? -1 : -c)
    }))
    .add("l", new KeyNode((vi: VI, c: number) => {
        vi.ed.moveH( (c == 0) ? 1 : c)
    }))
    .add("0", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        vi.ed.moveTo(curL.start)
    }))
    .add("^", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        vi.ed.moveTo(curL.start)

        const nextW = vi.ed.nextWord();
        if (typeof(nextW) !== "undefined") {
            vi.ed.moveTo(nextW.start);
        }
    }))
    .add("$", new KeyNode((vi: VI, c: number) => {
        if ( c > 1 ) { vi.ed.moveV(c-1) }
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        vi.ed.moveTo(curL.end);
    }))
    .add("G", new KeyNode((vi: VI, c: number) => {
        if ( c === 0 ) {
            const l = vi.ed.line(vi.ed.EOF());
            if (typeof(l) !== "undefined") vi.ed.moveTo(l.start);
        } else {
            const l = vi.ed.lineV(c-1, vi.ed.line(0));
            if (typeof(l) !== "undefined") vi.ed.moveTo(l.start);
        }
    }))
    .add("w", new KeyNode((vi: VI, c: number) => {
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            const nextW = vi.ed.nextWord();
            if (typeof(nextW) !== "undefined") {
                vi.ed.moveTo(nextW.start);
            }
        }
    }))
    .add("i", new KeyNode((vi: VI, c: number) => {
        vi.setMode("INSERT")
    }))
    .add("a", new KeyNode((vi: VI, c: number) => {
        vi.ed.moveH(1);
        vi.setMode("INSERT");
    }))
    .add("A", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        vi.ed.moveTo(curL.end);
        vi.setMode("INSERT");
    }))
    .add(">", new KeyNode((vi: VI, c: number) => {
        let pos = vi.ed.cursor();
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            const curL = vi.ed.line(pos);
            if (typeof(curL) === "undefined") return;
            vi.ed.insert(" ".repeat(vi.shiftwidth), curL.start);

            if (curL.end === vi.ed.EOF()) break;
            pos = curL.end + 1;
        }

        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;

        let ii = curL.start;
        for (; ii < curL.end; ii++) {
            const c = vi.ed.text().charAt(ii);
            if (!/(\p{Z})/u.test(c)) break;
        }
        vi.ed.moveTo(ii);
    }))
    .add("<", new KeyNode((vi: VI, c: number) => {
        let pos = vi.ed.cursor();
        for (let i = (c === 0)? 1 : c; i > 0; i--) {
            const curL = vi.ed.line(pos);
            if (typeof(curL) === "undefined") return;

            let ii = curL.start;
            for (; ii < curL.end; ii++) {
                const c = vi.ed.text().charAt(ii);
                if (!/(\p{Z})/u.test(c)) break;
            }
            const sp = (ii - curL.start);
            if (sp > vi.shiftwidth) {
                vi.ed.delete(curL.start, curL.start + vi.shiftwidth);
                vi.ed.moveTo(ii - vi.shiftwidth);
            } else if (sp > 0) {
                vi.ed.delete(curL.start, curL.start + sp);
                vi.ed.moveTo(ii - sp);
            }

            if (curL.end === vi.ed.EOF()) break;
            pos = curL.end + 1;
        }
    }))
    .add("J", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        if (curL.end === vi.ed.EOF()) return;

        vi.ed.delete(curL.end);
        while (vi.ed.text().charAt(curL.end) === " ") {
            vi.ed.delete(curL.end)
        }

        vi.ed.insert(" ", curL.end);
        vi.ed.moveTo(curL.end)
    }))
    .add("o", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;

        if (vi.ed.text().charAt(curL.end) !== "\n") vi.ed.insert("\n", curL.end + 1);
        vi.ed.insert("\n", curL.end + 1);
        vi.ed.moveV(1);
        vi.setMode("INSERT");
    }))
    .add("O", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;

        vi.ed.insert("\n", curL.start);
        vi.setMode("INSERT");
    }))
    .add("x", new KeyNode((vi: VI, c: number) => {
        const start = vi.ed.cursor();
        let end = ( c > 0 ) ? start + c : start + 1; 

        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        if ( end > curL.end) { end = curL.end; }
        
        vi.copy(start, end);
        vi.ed.delete(start, end);
    }))
    .add("D", new KeyNode((vi: VI, c: number) => {
        const curL = vi.ed.line();
        if (typeof(curL) === "undefined") return;
        const pos = vi.ed.cursor();
        vi.copy(pos, curL.end);
        vi.ed.delete(pos, curL.end);
    }))
    .add("d", new KeyNode()
        .add("d", new KeyNode((vi: VI, c: number) => {
            const curL = vi.ed.line();
            if (typeof(curL) === "undefined") return;
            const delL = (c > 1) ? vi.ed.lineV(c-1) : curL;
            if (typeof(delL) === "undefined") return;

            if (delL.end === vi.ed.EOF()) {
                if (delL.start > 0) delL.start = delL.start - 1;
                vi.copy(delL.start, delL.end);
                vi.ed.delete(delL.start, delL.end);

                const newCurL = vi.ed.line(delL.start)
                if (typeof(newCurL) === "undefined") return;
                vi.ed.moveTo(newCurL.start);
            } else {
                vi.copy(delL.start, delL.end + 1);
                vi.ed.delete(delL.start, delL.end + 1);
                vi.ed.moveTo(delL.start);
            }
        }))
        .add("w", new KeyNode((vi: VI, c: number) => {
            const start = vi.ed.cursor();
            let end = -1;
            for (let i = (c === 0)? 1 : c; i > 0; i--) {
                const nextW = vi.ed.nextWord();
                if (typeof(nextW) === "undefined") break;
                end = nextW.start;
            }
            if (end === -1) return;
            vi.copy(start, end);
            vi.ed.delete(start, end);
        }))
    )
    .add("y", new KeyNode()
        .add("y", new KeyNode((vi: VI, c: number) => {
            const curL = vi.ed.line();
            if (typeof(curL) === "undefined") return;
            const endL = (c > 1) ? vi.ed.lineV(c-1) : curL;
            if (typeof(endL) === "undefined") return;
            vi.copy(curL.start, endL.end+1);
        }))
        .add("w", new KeyNode((vi: VI, c: number) => {
            const start = vi.ed.cursor();
            let end = -1;
            for (let i = (c === 0)? 1 : c; i > 0; i--) {
                const nextW = vi.ed.nextWord();
                if (typeof(nextW) === "undefined") break;
                end = nextW.start;
            }
            if (end === -1) return;
            vi.copy(start, end);
        }))
    )
    .add("p", new KeyNode((vi: VI, c: number) => {
        if (vi.clipboard[vi.clipboard.length - 1] === "\n") {
            const curL = vi.ed.line();
            if (typeof(curL) === "undefined") return;
            const start = curL.end + 1;
            if (start >= vi.ed.EOF()) {
                vi.ed.insert("\n", start);
            }
            vi.paste(start);
        } else {
            vi.paste(vi.ed.cursor() + 1);
        }
    }))
    .add("P", new KeyNode((vi: VI, c: number) => {
        if (vi.clipboard[vi.clipboard.length - 1] === "\n") {
            const curL = vi.ed.line();
            if (typeof(curL) === "undefined") return;
            vi.paste(curL.start);
            
        } else {
            const start = vi.ed.cursor();
            vi.paste((start === 0) ? start : start - 1);
        }
    }))
    .add("u", new KeyNode((vi: VI, c: number): Action => {
        if (vi.ctree.lastCmd && vi.ctree.lastCmd.key === "u_undo") {
            vi.ed.redo();
            return {key: "u_redo", action: (vi: VI, c: number) => {vi.ed.undo()}};
        }

        vi.ed.undo();
        return {key: "u_undo", action: (vi: VI, c: number) => {vi.ed.undo()}};
    }))
    .add(".", new KeyNode((vi: VI, c: number) => {
        if (vi.ctree.lastCmd) vi.ctree.lastCmd.action(vi, c);
        return vi.ctree.lastCmd;
    }));
