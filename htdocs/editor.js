var editor = (function() {
    const form = document.querySelector("form");
    const content = form.querySelector("textarea[name=content]");
    const submit = form.querySelector("input[type=submit]");
    const status = form.querySelector("#status");

    async function save(filepath) {
        filepath = ((typeof(filepath) === "undefined") || (filepath === "")) ? form.action : filepath;
        try {
            let response = await fetch(filepath, {
                body: new FormData(form),
                method: form.method
            });

            if (!response.ok) {
                let err = await response.text();
                submit.style = "color:red";
                throw err;
            } else {
                submit.disabled = true;
                submit.style = "";
                if (filepath !== form.action) window.open(filepath);
            }
        } catch(err) {
            throw err;
        }
    }

    function open(filepath) {
        if (filepath.startsWith('/')) {
            window.location.pathname = filepath;
        } else {
            let lastPath = window.location.pathname.lastIndexOf("/");
            let dirname = (lastPath > 0) ? window.location.pathname.slice(0, lastPath+1) : "";
            window.location.pathname = dirname + filepath;
        }
    }

    function quit() {
        let lastPath = window.location.pathname.lastIndexOf("/");

        let dirname = (lastPath > 0) ? window.location.pathname.slice(0, lastPath) : "";
        window.location.pathname = dirname;
    }

    submit.disabled = true;
    content.addEventListener("input", function() {
        submit.disabled = false;
    });

    form.addEventListener("submit", function(event) {
        event.preventDefault();
        save();
        content.focus();
    });

    return {
        content: content,
        status: status,
        save: save,
        quit: quit,
        open: open,
    };
}());


window.onload = function() {
    var vi = new VI(editor.content, editor.status);
    vi.open = editor.open;
    vi.save = editor.save;
    vi.quit = editor.quit;
};
