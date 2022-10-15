const editorForm = document.getElementById('editorForm')
const editor = document.getElementById('editor')
const saveBtn = document.getElementById('saveBtn')
const statusMsg = document.getElementById('statusMsg')

editor.addEventListener('input', function() {
    saveBtn.disabled = false;
});

async function save_editor_content() {
    try {
        let response = await fetch(editorForm.action, {
            method: editorForm.method,
            body: new FormData(editorForm),
        });

        if (!response.ok) {
            saveBtn.style = "color:red";

            let errMsg = await response.text();
            statusMsg.innerHTML = 'Error: ' + errMsg;
        } else {
            saveBtn.disabled = true;
            saveBtn.style = "";
        }

    } catch (error) {
        statusMsg.innerHTML = 'Error: ' + error;
    }
}

editorForm.addEventListener('submit', function(event) {
    event.preventDefault();
    save_editor_content();
    editor.focus();
});

function save_content_from_vim(vim, cdata) {
    vim.log('save editor content to '+editorForm.action);
    save_editor_content();
}

window.onload = function() {
    const vim = new VIM();

    vim.on_set_mode = function(vi){
        statusMsg.innerHTML = (this.m_mode !== COMMAND) ? '-- ' + vi.m_mode + ' --' : '';
    }

    vim.m_ctrees[COMMAND].set_choice(':', node()
        .set_choice('w', node()
            .set_choice('<Enter>', node({action: save_content_from_vim}))
        )
    );

    vim.attach_to(editor);
    saveBtn.disabled = true;
}
