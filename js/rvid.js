window.onload = function(){
    var editor = document.getElementById('editor')
    if (editor != null) {
        var vim = new VIM()

        var status = document.getElementById('status')
        if (status != null) {
            vim.on_set_mode = function(vi){
                status.innerHTML = (this.m_mode !== COMMAND) ? '-- ' + vi.m_mode + ' --' : ''
            }
        }

        vim.attach_to(editor)
    }
}
