window.onload = function() {
    var editor = document.getElementById('editor')
    if (editor != null) {
        var vim = new VIM();

        const statusMsg = document.getElementById('statusMsg');
        if (statusMsg != null) {
            vim.on_set_mode = function(vi){
                statusMsg.innerHTML = (this.m_mode !== COMMAND) ? '-- ' + vi.m_mode + ' --' : '';
            }
        }

        vim.attach_to(editor);
        editor.focus();
    }
}
