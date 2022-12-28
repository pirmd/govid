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
    };
}());


window.onload = function() {
    var vi = new VI(editor.content, editor.status);
    vi.save = editor.save;
};
