import  ClipboardJS from "clipboard"
function init_clipboard() {
    let clipboard = new ClipboardJS('.btn');
    clipboard.on('success', function (e) {
        // console.info('Action:', e.action);
        // console.info('Text:', e.text);
        // console.info('Trigger:', e.trigger);
        e.clearSelection();
    });
    clipboard.on('error', function (e) {
        console.error('Action:', e.action);
        console.error('Trigger:', e.trigger);
    });
}
export default {
    init_clipboard
}