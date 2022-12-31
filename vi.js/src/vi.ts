type Mode = "COMMAND" | "INSERT" | "EXECUTE";

class VI {
    shiftwidth = 4;
    save = () => { throw new Error("Save is not implemented"); };
    quit = () => { throw new Error("Quit is not implemented"); };

    mode: Mode = "COMMAND";
    ctree: CommandTree = new CommandTree(this);
    ed: Editor;
    ex: HTMLInputElement;
    clipboard = "";
    lastSearch = "";

    constructor(textarea: HTMLTextAreaElement, exInput: HTMLInputElement) {
        textarea.onkeydown = this.onkeydown.bind(this);
        this.ed = new Editor(textarea);

        exInput.onkeydown = this.onkeydown.bind(this);
        this.ex = exInput;
    }

    setMode(mode: Mode) {
        if (mode === "EXECUTE") {
            this.ex.disabled = false;
            this.ex.focus();
        } else {
            this.ex.disabled = true;
            this.ed.focus();
        }

        if (mode === "COMMAND") this.ctree.reset();

        if (this.mode === "INSERT") this.ex.value = "";
        if (mode === "INSERT") this.ex.value = "-- INSERT --";

        this.mode = mode;
    }

    onkeydown(event: KeyboardEvent) {
        let k = event.key;
        if (event.ctrlKey) k = "Ctrl-" + k;
        if (this.handleKey(k)) event.preventDefault();
    }

    handleKey(k: string) {
        switch (true) {
            case (k === "Escape"):
                this.setMode("COMMAND");
                return true;

            case (this.mode === "INSERT"):
                if (k === "Enter") { //auto indent
                    this.ed.insert("\n");
                    this.ed.moveV(1);
                    this.ed.indent();
                    return true;
                }
                return false;

            case (this.mode == "EXECUTE"):
                if (k === "Enter") {
                    this.execute();
                    this.setMode("COMMAND");
                    return true;
                }
                return false;

            default: // "COMMAND"
                if ((k === ":") || (k === "/") || (k === "?")) {
                    this.setMode("EXECUTE");
                    return false;
                }
                this.ctree.handleKey(k);
                return true;
        }
    }

    execute() {
        const cmdline = this.ex.value;
        this.ex.value = "";

        switch (true) {
            case (cmdline === ":q"):
                this.quit();
            return;

            case (cmdline === ":w"):
                try {
                    this.save();
                } catch(err) {
                    if (err instanceof Error) this.err(err.message);
                }
                return;

            case (cmdline === ":wq"):
                try {
                    this.save();
                } catch(err) {
                    if (err instanceof Error) this.err(err.message);
                    return;
                }

                this.quit();
                return;

            case ((cmdline.startsWith("/")) || (cmdline.startsWith("?"))): {
                if (cmdline.length > 1) this.lastSearch = cmdline.slice(1);

                if (this.lastSearch === "") {
                    this.err("No previous search pattern");
                    return;
                }

                let pos = -1;
                if (cmdline.startsWith("/")) {
                    pos = this.ed.search(this.lastSearch);
                } else {
                    pos = this.ed.searchBackwards(this.lastSearch);
                }

                if (pos === -1) {
                    this.err("Pattern not found");
                    return;
                }

                this.ed.moveTo(pos);
                return;
            }

            default:
                this.err("Not an editor command");
        }
    }

    copy(start?: number, end?: number) {
        start = (typeof(start) === "undefined") ? this.ed.cursor() : start;
        end = (typeof(end) === "undefined") ? start + 1 : end;

        this.clipboard = this.ed.text().slice(start, end);
    }

    paste(start?: number, end?: number) {
        this.ed.insert(this.clipboard, start, end);
    }

    err(msg: string) {
        this.ex.value = msg;
        setTimeout(() => this.ex.value = "", 1000);
    }
}
