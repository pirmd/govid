const editorForm = document.getElementById('editorForm')
const editor = document.getElementById('editor')
const saveBtn = document.getElementById('saveBtn')
const statusMsg = document.getElementById('statusMsg')

editor.addEventListener('input', function() {
    saveBtn.disabled = false;
});

editorForm.addEventListener('submit', async function(event) {
    event.preventDefault();

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

    editor.focus();
});

window.onload = function() {
    const vim = new VIM();

    vim.on_set_mode = function(vi){
        statusMsg.innerHTML = (this.m_mode !== COMMAND) ? '-- ' + vi.m_mode + ' --' : '';
    }

    vim.attach_to(editor);
    editor.focus();
}
