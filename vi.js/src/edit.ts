type StartEnd = { start: number; end: number };
type Snapshot = { pos: number; text: string };

class Editor {
    textarea: HTMLTextAreaElement;
    history: Array<Snapshot> = [];
    historyRedo: Array<Snapshot> = [];
    maxHistory = 64;

    constructor(textarea: HTMLTextAreaElement) {
        this.textarea = textarea;
        this.snapshot();
    }

    text(): string {
        return this.textarea.value;
    }

    EOF(): number {
        return this.text().length - 1;
    }

    focus() {
        this.textarea.focus();
    }

    cursor(): number {
        return this.textarea.selectionStart;
    }

    moveTo(pos: number) {
        this.textarea.setSelectionRange(pos, pos);
    }

    moveH(di: number) {
        let pos = this.cursor() + di;

        const curL = this.line();
        if (typeof(curL) === "undefined") return;

        pos = (pos < curL.start) ? curL.start : pos;
        pos = (pos > curL.end) ? curL.end : pos;

        this.moveTo(pos);
    }

    moveV(dj: number) {
        const curL = this.line();
        if (typeof(curL) === "undefined") return;

        const targetL = this.lineV(dj, curL);
        if (typeof(targetL) === "undefined") return;

        let pos = targetL.start + (this.cursor() - curL.start);
        pos = (pos > targetL.end) ? targetL.end : pos;

        this.moveTo(pos);
    }

    line(pos?: number): StartEnd | undefined {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;

        if (pos < 0 || pos > this.EOF()) {
            return undefined;
        }

        const start = (this.text().charAt(pos) === "\n") ?
            (pos > 0) ? this.text().lastIndexOf("\n", pos-1) + 1 : 0 :
            this.text().lastIndexOf("\n", pos) + 1;

        const end = this.text().indexOf("\n", pos);

        return {
            'start': start,
            'end': (end === -1) ? this.EOF() : end,
        };
    }

    lineV(dj: number, fromL?: StartEnd): StartEnd | undefined {
        let nextL = (typeof(fromL) === "undefined") ? this.line() : fromL;
        if (typeof(nextL) === "undefined") return undefined;

        for (let j = dj; j !== 0; (j > 0) ? j-- : j++) {
            nextL = (j > 0) ? this.line(nextL.end+1) : this.line(nextL.start-1);
            if (typeof(nextL) === "undefined") break;
        }

        return nextL;
    }

    word(pos?: number): StartEnd | undefined {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;
        const c = this.text().charAt(pos);
        if (/(\p{Z}|\n)/u.test(c)) return undefined;
        if (/\p{P}/u.test(c)) return {start: pos, end: pos};

        let end = pos;
        for (end = pos+1; end <= this.EOF(); end++) {
            const c = this.text().charAt(end);
            if (/(\n|\p{Z}|\p{P})/u.test(c)) {
                end--;
                break;
            }
        }

        let start = pos;
        for (start = pos-1; start > 0; start--) {
            const c = this.text().charAt(start);
            if (/(\n|\p{Z}|\p{P})/u.test(c)) {
                start++;
                break;
            }
        }

        return {start: start, end: end};
    }

    nextWord(pos?: number): StartEnd | undefined {
        pos = (typeof(pos) === "undefined") ? this.cursor() : pos;
        const curW = this.word(pos);
        const curP = (typeof(curW) === "undefined") ? pos : curW.end + 1;

        let nextW = undefined
        for (let p = curP; p <= this.EOF(); p++) {
            nextW = this.word(p);
            if (typeof(nextW) !== "undefined") break;
        }

        return nextW;
    }
    
    insert(txt: string, start?: number, end?: number) {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        end = (typeof(end) === "undefined") ? start : end;

        this.snapshot();
        this.textarea.setRangeText(txt, start, end);
        this.textarea.dispatchEvent(new Event("input"));
    }

    delete(start: number, end?: number) {
        if (typeof(end) === "undefined") {
            end = (start < this.EOF()) ? start + 1: start;
        }

        this.insert("", start, end);
    }

    search(value: string, start?: number): number {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        return this.textarea.value.indexOf(value, start);
    }

    searchBackwards(value: string, start?: number): number {
        start = (typeof(start) === "undefined") ? this.cursor() : start;
        return this.textarea.value.lastIndexOf(value, start);
    }

    indent() {
        const curL = this.line();
        if (typeof(curL) === "undefined") return;

        let i = curL.start;
        for (; i < curL.end; i++) {
            const c = this.text().charAt(i);
            if (!/(\p{Z})/u.test(c)) break;
        }
        const curSP = (i - curL.start);

        const prevL = this.line(curL.start - 1);
        if (typeof(prevL) === "undefined") return;

        let j = prevL.start;
        for (; j < prevL.end; j++) {
            const c = this.text().charAt(j);
            if (!/(\p{Z})/u.test(c)) break;
        }
        const prevSP = (j - prevL.start);

        if (curSP < prevSP) {
            const sp = prevSP - curSP;
            this.insert(" ".repeat(sp), curL.start);
            this.moveTo(i + sp);
        } else if (curSP > prevSP) {
            const sp = curSP - prevSP
            this.delete(curL.start, curL.start + sp);
            this.moveTo(i - sp);
        }
    }

    snapshot() {
        if (this.history.length > 0 && (this.text() === this.history[this.history.length - 1].text)) return;

        this.history.push({
            text: this.text(),
            pos:  this.cursor(),
        });

        if (this.history.length > this.maxHistory) {
            this.history.shift();
        }
    }

    undo() {
        if (this.history.length > 0) {
            this.historyRedo.push({
                text: this.text(),
                pos:  this.cursor(),
            });
            if (this.historyRedo.length > this.maxHistory) {
                this.historyRedo.shift();
            }

            const u = this.history.pop();
            if (typeof(u) === "undefined") return;

            this.textarea.value = u.text;
            this.moveTo(u.pos);
        }
    }

    redo() {
        if (this.historyRedo.length > 0) {
            this.snapshot();

            const u = this.historyRedo.pop();
            if (typeof(u) === "undefined") return;

            this.textarea.value = u.text;
            this.moveTo(u.pos);
        }
    }
}
