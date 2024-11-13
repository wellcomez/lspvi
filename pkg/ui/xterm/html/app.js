/**
 * Copyright 2024 wellcomez
 * SPDX-License-Identifier: gplv3
 */

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
    return ["md"].includes(ext)
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




const MINIMUM_COLS = 2;
const MINIMUM_ROWS = 1;
var ws_sendTextData
class fullscreen_check {
    constructor(term) {
        this.term = term
        // fit.call(this)
    }
    dims = () => {
        if (!this.term) {
            return undefined;
        }

        if (!this.term.element || !this.term.element.parentElement) {
            return undefined;
        }

        // TODO: Remove reliance on private API
        const core = this.term._core;
        const dims = core._renderService.dimensions;

        if (dims.css.cell.width === 0 || dims.css.cell.height === 0) {
            return undefined;
        }
        return dims

    }
    scrollbarWidth = () => {
        const scrollbarWidth = (this.term.options.scrollback === 0
            ? 0
            : (this.term.options.overviewRuler?.width || 1));
        return scrollbarWidth
    }
    resize = (call) => {
        this.term.element.style.height = (window.innerHeight - 50) + "px"
        this.term.element.style.width = window.innerWidth + "px"
        let dims = this.check()
        // console.log("col-row", ss)
        if (dims != undefined) {
            this.term.resize(dims.cols, dims.rows);
            if (call)
                resizecall()
        }
    }
    check = () => {
        const dims = this.dims()
        const scrollbarWidth = this.scrollbarWidth()
        const parentElementStyle = window.getComputedStyle(this.term.element.parentElement);
        const parentElementHeight = parseInt(parentElementStyle.getPropertyValue('height'));
        const parentElementWidth = Math.max(0, parseInt(parentElementStyle.getPropertyValue('width')));
        const elementStyle = window.getComputedStyle(this.term.element);
        const elementPadding = {
            top: parseInt(elementStyle.getPropertyValue('padding-top')),
            bottom: parseInt(elementStyle.getPropertyValue('padding-bottom')),
            right: parseInt(elementStyle.getPropertyValue('padding-right')),
            left: parseInt(elementStyle.getPropertyValue('padding-left'))
        };
        const elementPaddingVer = elementPadding.top + elementPadding.bottom;
        const elementPaddingHor = elementPadding.right + elementPadding.left;
        const availableHeight = parentElementHeight - elementPaddingVer;
        const availableWidth = parentElementWidth - elementPaddingHor - scrollbarWidth;
        return this.fit(availableWidth, availableHeight, dims)
    }
    fit = (availableWidth, availableHeight, dims) => {
        const geometry = {
            cols: Math.max(MINIMUM_COLS, Math.floor(availableWidth / dims.css.cell.width)),
            rows: Math.max(MINIMUM_ROWS, Math.floor(availableHeight / dims.css.cell.height))
        };
        return geometry;
    }
}
md_init = () => {
    const mdIt = markdownit({
        html: true,
        linkify: true,
        typographer: true,
        highlight: function (str, lang) {
            if (lang == "plantuml") {
                return "<div class=\"plantuml\">" +
                    mdIt.render(str) +
                    "</div>"
            }
            if (lang && hljs.getLanguage(lang)) {
                try {
                    return '<pre><code class="hljs">' +
                        lang +
                        "<br>" +
                        "<br>"
                        + hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
                        '</code></pre>';
                } catch (__) { }
            }
            return '<pre><code class="hljs">' + mdIt.utils.escapeHtml(str) + '</code></pre>';
        }
    })
    class md {
        constructor() {
            this.md = mdIt
        }
        on_wheel = (evt) => {
            let div = document.getElementsByClassName("md")[0]
            let top = div.scrollTop + evt.deltaY
            div.scroll(div.scrollLeft, top)
        }
        render = (url) => {
            axios.get(url, { responseType: "text" }).then((resp) => {
                let ss = this.md.render(resp.data)
                let div = document.getElementsByClassName("md")[0]
                div.innerHTML = ss
                document.getElementsByClassName("md")[0].addEventListener("click", (event) => {
                    console.log(event)
                })
            });
        }
    }
    return new md()
}
const check_event_in_element = (classname, event) => {
    var targetDiv = document.getElementsByClassName(classname)[0];
    var rect = targetDiv.getBoundingClientRect();
    let yes = false
    if (targetDiv) {
        if (event.clientX >= rect.left && event.clientX <= rect.right &&
            event.clientY >= rect.top && event.clientY <= rect.bottom) {
            yes = true
        } else {
            yes = false
        }
    }

    return { el: targetDiv, yes: yes }
}
app_init = () => {
    let md = md_init()
    let app = new Vue({
        el: '#app',
        data: {
            message: 'Hello Vue!',
            isVisible: false,
            isVisibleMd: false,
            imageurl: "",
        },
        methods: {
            on_forward(event) {
                let { bubbles, target,
                    cancelable,
                    view,
                    clientX,
                    clientY } = event
                const { el, yes } = check_event_in_element("md", event)
                if (yes == false) {
                    return false
                } else if (el != target) {
                    event.preventDefault()
                    // const clickevent = new PointerEvent(event.type, event)
                    // el.dispatchEvent(clickevent)
                    return true
                } else {
                    return true
                }
            },
            is_hide() {
                if (this.isVisibleMd == undefined && this.isVisible == undefined) {
                    return true
                } else {
                    return false
                }
            },
            onhide() {
                this.set_visible({})
            },
            on_mouse(event) {
                if (event.type == "click" && this.isVisible) {
                    this.onhide()
                    return true
                }
                if (this.isVisibleMd) {
                    if (this.on_forward(event) == false) {
                        if (event.type == "click") {
                            this.onhide()
                        }
                    }
                    return true
                }
                return false
            },
            on_wheel(evt) {
                if (this.isVisibleMd) {
                    md.on_wheel(evt)
                    return true
                }
                else if (this.isVisible) {
                    return true
                }
            },
            set_visible(a) {
                let { isVisibleMd, isVisible } = a
                this.isVisible = isVisible
                this.isVisibleMd = isVisibleMd
            },
            popimage(image) {
                this.set_visible({ isVisible: true })
                this.imageurl = image
            },
            popmd(image) {
                let u = image
                app.set_visible({ isVisibleMd: true })
                md.render(u)
            }
        }
    })

    wheel = (evt) => {
        console.log(evt)
    }
    document.addEventListener("wheel", wheel)
    const div = document.getElementById('terminal')
    if (div) {
        div.addEventListener("wheel", wheel)
    }
    return app
}
class RemoteTermStatus {
    constructor() {
    }
}
class RemoteConn {
    constructor(socket) {
        this.socket = socket
    }
}
class LocalTerm {
    constructor(term, conn) {
        this.term = term
        this.conn = conn
        this.prompt = "bash# "
        this.term.clear()
        this.newline();
        let lsp = (cmd) => {
            if (cmd.indexOf("lspvi") == 0) {
                this.conn.start_lspvi(cmd)
                return true
            }
        }
        lsp = lsp.bind(this)
        this.local_cmd_matcher = [lsp]
    }
    newline() {
        this.term.write(this.prompt);
    }
    // Assuming you have an xterm.js instance created as 'terminal'

    getCurrentLineText = () => {
        // Get the cursor position
        const cursorY = this.term.buffer.active.cursorY;

        // Get the text of the current line
        const lineText = this.term.buffer.active.getLine(cursorY).translateToString().trim();

        return lineText;
    };

    // Usage example
    handleCommand(cmdline) {
        //let args = cmdline
        let matched = false;
        this.local_cmd_matcher.forEach(element => {
            if (matched) return;
            if (element(cmdline)) {
                matched = true;
                return;
            }
        });
        return false;
    }
    ondata(data) {
        const { term } = this
        const currentBuffer = term.buffer.active;
        if (data == '\r') {
            let line = this.getCurrentLineText()
            if (line.indexOf(this.prompt) == 0) {
                if (this.handleCommand(line.substring(this.prompt.length))) {

                    return
                }

            }
            this.term.write('\r\n')
            this.newline()
            return
        } else if (data === '\x7F') { // Delete key or similar
            if (currentBuffer.cursorX > this.prompt.length) {
                term.write('\x08'); // Backspace
                term.write(' ');    // Replace with space
                term.write('\x08'); // Backspace again to move cursor back
            }
        }
        this.term.write(data)
    }
}






function handle_copy_data(data) {
    let text = data.SelectedString;
    let txt = document.getElementById("bar");
    txt.innerText = text;
    let btn = document.getElementById("clip");
    btn.click();
    return text
}




class Term {
    constructor(term) {
        this.set_term_ui(term)

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
            this.term.options.disableStdin = true
            app.popmd(data.Filename);
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

var termobj = new Term()
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
    let term = new Terminal({
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
    let f = new fullscreen_check(term)
    f.resize(false)
    fit.fit()
    termobj.fit = fit
    term.focus()

    const resize_handle = function (evt) {
        let f = new fullscreen_check(term);
        f.resize();
        fit.fit();
    };
    termobj.resize_handle = resize_handle
    window.addEventListener('resize', resize_handle)
}

const socket_int = (term_obj, app, start_lspvi) => {
    let localhost = window.location.host
    let wsproto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    let socket = new WebSocket(wsproto + '://' + localhost + '/ws');
    let conn = new RemoteConn(socket)
    term_obj.conn = conn;
    const sendTextData = (data) => {
        if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(data));
            // console.log('Sent to server:', data);
        } else {
            console.error('WebSocket connection is not open.');
        }
    }
    term_obj.sendTextData = sendTextData
    ws_sendTextData = sendTextData

    socket.onopen = function (event) {
        start_lspvi();
    };


    socket.binaryType = "blob";
    socket.onmessage = function incoming(evt) {
        try {
            let reader = new FileReader();
            reader.readAsArrayBuffer(evt.data);
            reader.addEventListener("loadend", function (e) {
                const buffer = new Uint8Array(e.target.result);  // arraybuffer object
                const message = msgpack5().decode(buffer);
                term_obj.handleMessage(message, app)
            });
        } catch (error) {
            console.error('Failed to decode data:', error);
        }

    };
    socket.onclose = function (event) {
        console.error("Connection closed");
    };




    conn.start_lspvi = start_lspvi.bind(conn)
}
const main = () => {
    let app = app_init()
    termobj.setapp(app)
    socket_int(termobj, app, (cmdline) => {
        termobj.start_xterm(true, cmdline)
    })

}
main()


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
