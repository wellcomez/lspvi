import Terminal from '@xterm/xterm';
import { WebglAddon } from '@xterm/addon-webgl';
import { FitAddon } from '@xterm/addon-fit';
import clip from './clipboard'
import full from './fullscreen_check'

function getFileExtension(filename) {
    const lastDotIndex = filename.lastIndexOf('.');
    if (lastDotIndex === -1) {
        return ''; // 没有扩展名
    }
    return filename.substring(lastDotIndex + 1);
}
function is_image(ext) {
    return ["jpg", "png", "gif", "jpeg", "bmp"].includes(ext)
}

function is_md(ext) {
    return ["md", "puml"].includes(ext)
}

let rows = 50
let cols = 80
const call_key = "key"
const call_term_stdout = "term"
const call_xterm_init = "init"
const call_resize = "resize"
const call_paste_data = "call_paste_data"
const call_redraw = "call_redraw"
const forward_call_refresh = "forward_call_refresh"
const lspvi_backend_start = "xterm_lspvi_start"

const backend_on_command = "call_term_command"
const backend_on_zoom = "zoom"
const backend_on_copy = "onselected"
const backend_on_openfile = "openfile"
// const backend_on_command = "call_term_command"


// const backend_on_command = "call_term_command"
class RemoteTermStatus {
    constructor() {
    }
}

class Term {
    constructor(term) {
        this.set_term_ui(term)
        clip.init_clipboard()
        // let clipboard = new ClipboardJS('.btn');
        // clipboard.on('success', function (e) {
        //     // console.info('Action:', e.action);
        //     // console.info('Text:', e.text);
        //     // console.info('Trigger:', e.trigger);
        //     e.clearSelection();
        // });
        // clipboard.on('error', function (e) {
        //     console.error('Action:', e.action);
        //     console.error('Trigger:', e.trigger);
        // });

        this.appstatus = new RemoteTermStatus()
    }
    setapp(app) {
        this.app = app
        let obj = this
        window.addEventListener("contextmenu", function (ev) {
            if (app.on_mouse(ev)) {
            } else {
                ev.preventDefault();
            }
        })
        document.addEventListener("mouseup", (ev) => {
            if (app.on_mouse(ev)) {
            }
        })
        document.addEventListener("mousedown", (ev) => {
            if (app.on_mouse(ev)) {
            }
        })
        document.addEventListener("mousemove", function (ev) {
            if (app.on_mouse(ev)) {
            }
        })
        document.addEventListener("click", function (ev) {
            if (app.on_mouse(ev)) {
            }
            if (app.is_hide()) {
                obj.term.options.disableStdin = false
            }
        })
    }
    paste_text = (data) => {
        let call = call_paste_data
        this.sendTextData({ call, data })
    }
    handleMessage = (data, app) => {
        let { Call, Output } = data
        let { term } = this
        if (Call == call_term_stdout) {
            term.write(Output)
        }
        if (this.handle_backend_command(Call, data, app)) {
            return
        }
    }
    start_xterm(start, cmdline) {
        term_init(this, this.app);
        if (start) {
            this.start_lspvi(cmdline);
        }

    }
    handle_command_zoom(data) {
        let zoomout = data.Zoom
        let fontsize = get_font_size();
        let { cols, rows } = this.term
        let zoomfactor = fontsize
        if (zoomout) {
            fontsize++
        } else {
            fontsize--
        }
        zoomfactor = fontsize / zoomfactor
        set_font_size(fontsize);
        let { term } = this
        term.options.fontSize = fontsize
        cols = Math.ceil(cols / zoomfactor)
        rows = Math.ceil(rows / zoomfactor)
        console.log("resize", term.cols, term.rows)
        term.resize(cols, rows)
        console.log("after resize", term.cols, term.rows)
        this.fit.fit()
        console.log("after resize fit", term.cols, term.rows)
        this.resizecall()
    }
    // reset(Zoom) {
    //     window.removeEventListener("resize", this.resizeListener);
    //     this.term.clear();
    //     this.term.dispose();
    //     // this.term.reset()
    //     let terminalContainer = document.getElementById('terminal');
    //     let parent = terminalContainer.parentElement;
    //     terminalContainer.remove();
    //     terminalContainer = document.createElement('div');
    //     terminalContainer.id = "terminal";
    //     parent.appendChild(terminalContainer);
    //     console.log("zoom", Zoom);
    //     this.start_xterm();
    //     this.sendTextData({ Call: call_redraw })
    // }
    handle_user_command(data) {
        switch (data.Command) {
            case "quit":
                this.appstatus.quit = true;
                this.on_remote_stop();
                break;
            default:
                return;
        }
    }
    handle_command_openfile(data, app) {
        console.log("openfile",
            data.Filename);
        let ext = getFileExtension(data.Filename);
        if (is_image(ext)) {
            this.term.options.disableStdin = true
            app.popimage(data.Filename);
        } else if (is_md(ext)) {
            const { PrjName } = data
            if (PrjName) {
                let host = window.location.host
                let url = "https://" + host + "/md/" + PrjName
                window.open(url);
                // this.term.options.disableStdin = true
                // app.popmd(url);
            } else {
                this.term.options.disableStdin = true
                app.popmd(data.Filename);
            }
        }
    }
    handle_backend_command(Call, data, app) {
        if (Call == backend_on_zoom) {
            this.handle_command_zoom(data);
        } else if (Call == backend_on_openfile) {
            this.handle_command_openfile(data, app);
        } else if (backend_on_copy == Call) {
            this.clipdata = handle_copy_data(data);
        } else if (Call == backend_on_command) {
            return this.handle_user_command(data);
        } else {
            return false
        }
        return true
    }
    start_lspvi(cmdline) {
        let { term } = this
        console.log("Connection opened");
        let call = call_xterm_init;
        let rows = term.rows, cols = term.cols;
        let host = window.location.host;
        this.Local = undefined
        this.status = {}
        if (!this.observer) {
            let termobj = this
            let callback = (el) => {
                let hasnew = false
                console.log(el)
                el.forEach((e) => {
                    let { addedNodes } = e
                    addedNodes.forEach((added) => {
                        if (added.classList.contains("focus")) {
                            hasnew = true
                        }
                    })
                })
                if (hasnew) {
                    let divs = document.querySelectorAll('.xterm')
                    divs.forEach((el) => {
                        if (el.classList.contains("focus") == false) {
                            el.remove()
                        }
                    })
                    termobj.resize_handle()
                }
            }

            const targetNode = document.getElementById('terminal')
            const config = { attributes: false, childList: true, subtree: false };
            const observer = new MutationObserver(callback.bind(this));
            observer.observe(targetNode, config);
            this.observer = observer
        }

        this.sendTextData({ call, cols, rows, host, cmdline });
        // let restart = true
        // this.status = { restart }
    }
    on_remote_stop() {
        let stop = true
        this.status = { stop }
        this.Local = new LocalTerm(this.term, this.conn)
    }
    onData(data) {
        if (this.status.stop) {
            this.Local.ondata(data);
        } else {
            let { term } = this
            let call = call_key;
            let rows = term.rows, cols = term.cols;
            this.sendTextData({ call, data, rows, cols });
        }
    }
    handle_key(ev) {
        // console.log(ev)
        if (this.Local) {
            // if (ev.code == "Backspace") {
            //     // term.write('\x7f')
            //     ev.preventDefault(); // 阻止默认行为
            //     // term.write('\b \b')
            //     term.write('\x08'); // Backspace
            //     return false;
            // }
            return true;
        }
        if (ev.ctrlKey) {
            if (ev.key == "=" || ev.key == "+") {
                this.handle_command_zoom({ Zoom: true })
                return false
            } else if (ev.key == "-") {
                this.handle_command_zoom({ Zoom: false })
                return false
            }
        }
        if (ev.key == "v" && ev.ctrlKey) {
            if (this.on_paste()) {
                return false;
            }
        }
        return true;
    };
    resizecall = () => {
        let { term } = this
        let call = call_resize
        let rows = term.rows, cols = term.cols
        this.sendTextData({ call, cols, rows })
    }
    set_term_ui(term) {
        this.term = term; this.status = {}
        if (term) {
            let a = this
            term.onResize((size) => {
                a.resizecall()
            })
        }
    }

    on_remote_inited() {
        let init = true
        this.status = { init }

    }
    on_paste() {
        if (this.clipdata != undefined) {
            this.paste_text(this.clipdata)
        }
        return true
    }
}

const term_init = (termobj, app) => {

    document.onkeydown = function (e) {
        e = e || window.event;//Get event
        if (!e.ctrlKey) return;
        let code = e.which || e.keyCode;//Get key code
        e.preventDefault();
        e.stopPropagation();
    };
    let fontSize = get_font_size();
    set_font_size(fontSize)
    let term = new Terminal.Terminal({
        allowProposedApi: true,
        cursorStyle: 'bar',  // 默认为块状光标
        allowTransparency: true,
        cursorBlink: false,
        cursorWidth: 4,
        rows: rows,
        cols: cols,
        vt200Mouse: true,
        x10Mouse: true,
        vt300Mouse: true,
        MouseEvent: true,
        fontSize: fontSize,
        // fontFamily: 'SymbolsNerdFontMono "Fira Code", courier-new, courier, monospace, "Powerline Extra Symbols"',
        // fontFamily: 'Hack, "Fira Code", monospace',
        // fontFamily: 'HackNerdFontMono,monospace'
        fontFamily: 'SymbolsNerdFontMono,courier-new, courier, monospace'

        // minimumContrastRatio: 1,
    });
    termobj.set_term_ui(term)
    let wl = new WebglAddon.WebglAddon()
    term.loadAddon(wl)
    let fit = new FitAddon.FitAddon()
    term.loadAddon(fit);
    term.onData(function (data) {
        termobj.onData(data);
    })
    const handle_key = ev => {
        return termobj.handle_key(ev)
    };
    term.attachCustomKeyEventHandler(handle_key)
    const handle_wheel = ev => {
        if (app && app.on_wheel(ev)) {
            return false;
        }
        // console.log(ev);
        return true;
    };
    term.attachCustomWheelEventHandler(handle_wheel)
    // term.onSelectionChange(() => {
    //     let word = term.getSelection()
    //     console.error("word:", word)
    // })
    old = ""
    term.open(document.getElementById('terminal'));
    // LoadLigaturesAddon();
    let f = new full.fullscreen_check(term)
    f.resize(false)
    fit.fit()
    termobj.fit = fit
    term.focus()

    const resize_handle = function (evt) {
        let f = new full.fullscreen_check(term);
        f.resize();
        fit.fit();
    };
    termobj.resize_handle = resize_handle
    window.addEventListener('resize', resize_handle)
}
function set_font_size(fontSize) {
    window.localStorage.setItem("fontsize", fontSize);
}
function get_font_size() {
    let fontSize = window.localStorage.getItem("fontsize");
    if (fontSize == undefined || fontSize == "undefined") {
        fontSize = 12;
    }
    return fontSize;
}
export default { Term }