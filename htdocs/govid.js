var editor = (function() {
    "use strict";

    const _form = document.querySelector("form");
    const _content = _form.querySelector("textarea[name=content]");
    const _submit = _form.querySelector("input[type=submit]");
    const _status = _form.querySelector("#status");

    function inform(message) {
        _status.innerHTML = message;
    }

    async function save() {
        try {
            let response = await fetch(_form.action, {
                body: new FormData(_form),
                method: _form.method
            });

            if (!response.ok) {
                let err = await response.text();

                _submit.style = "color:red";
                inform("Error: " + err);
            } else {
                _submit.disabled = true;
                _submit.style = "";
            }
        } catch(error) {
            inform("Error: " + error);
        }
    }

    _submit.disabled = true;
    _content.addEventListener("input", function() {
        _submit.disabled = false;
    });

    _form.addEventListener("submit", function(event) {
        event.preventDefault();
        save();
        _content.focus();
    });

    return {
        content: _content,
        inform: inform,
        save: save
    };
}());


window.onload = function() {
    const vim = new VIM();
    vim.on_set_mode = function(vi) {
        editor.inform(
            vi.m_mode !== COMMAND ? "-- " + vi.m_mode + " --" : ""
        );
    };
    vim.m_ctrees[COMMAND].set_choice(":", node()
        .set_choice("w", node()
            .set_choice("<Enter>", node({action: function(vim, cdata) {
                editor.save();
            }}))
        )
    );
    vim.attach_to(editor.content);
};
