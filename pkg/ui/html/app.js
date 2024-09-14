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

const call_on_copy = "onselected"
const MINIMUM_COLS = 2;
const MINIMUM_ROWS = 1;
var ws_sendTextData
class fullscreen_check {
    constructor(term) {
        this._terminal = term
        // fit.call(this)
    }
    dims = () => {
        if (!this._terminal) {
            return undefined;
        }

        if (!this._terminal.element || !this._terminal.element.parentElement) {
            return undefined;
        }

        // TODO: Remove reliance on private API
        const core = this._terminal._core;
        const dims = core._renderService.dimensions;

        if (dims.css.cell.width === 0 || dims.css.cell.height === 0) {
            return undefined;
        }
        return dims

    }
    scrollbarWidth = () => {
        const scrollbarWidth = (this._terminal.options.scrollback === 0
            ? 0
            : (this._terminal.options.overviewRuler?.width || 1));
        return scrollbarWidth
    }
    resize = (call) => {
        this._terminal.element.style.height = (window.innerHeight - 50) + "px"
        this._terminal.element.style.width = window.innerWidth + "px"
        let dims = this.check()
        // console.log("col-row", ss)
        if (dims != undefined) {
            this._terminal.resize(dims.cols, dims.rows);
            if (call)
                resizecall()
        }
    }
    check = () => {
        const dims = this.dims()
        const scrollbarWidth = this.scrollbarWidth()
        const parentElementStyle = window.getComputedStyle(this._terminal.element.parentElement);
        const parentElementHeight = parseInt(parentElementStyle.getPropertyValue('height'));
        const parentElementWidth = Math.max(0, parseInt(parentElementStyle.getPropertyValue('width')));
        const elementStyle = window.getComputedStyle(this._terminal.element);
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
            });
        }
    }
    return new md()
}
app_init = () => {
    var md = md_init()
    let app = new Vue({
        el: '#app',
        data: {
            message: 'Hello Vue!',
            isVisible: false,
            isVisibleMd: false,
            imageurl: "",
        },
        methods: {
            onhide() {
                this.set_visible({})
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
                var { isVisibleMd, isVisible } = a
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
    document.addEventListener("click", function () {
        app.onhide()
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
const term_init = (app) => {
    window.addEventListener("contextmenu", function (e) {
        e.preventDefault();
    })
    document.onkeydown = function (e) {
        e = e || window.event;//Get event
        if (!e.ctrlKey) return;
        var code = e.which || e.keyCode;//Get key code
        e.preventDefault();
        e.stopPropagation();
    };
    var fontSize = get_font_size();
    set_font_size(fontSize)
    var term = new Terminal({
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
    var wl = new WebglAddon.WebglAddon()
    term.loadAddon(wl)
    var fit = new FitAddon.FitAddon()
    term.loadAddon(fit);
    // const imageAddon = new ImageAddon.ImageAddon(customSettings);
    // terminal.loadAddon(imageAddon);
    term.onData(function (data) {
        let call = "key"
        let rows = term.rows, cols = term.cols
        ws_sendTextData({ call, data, rows, cols })
    })
    term.attachCustomKeyEventHandler(ev => {
        console.log(ev)
        return true;
    })
    term.attachCustomWheelEventHandler(ev => {
        if (app && app.on_wheel(ev)) {
            return false
        }
        console.log(ev)
        return true;
    })
    old = ""
    term.open(document.getElementById('terminal'));
    // LoadLigaturesAddon();
    let f = new fullscreen_check(term)
    f.resize(false)
    fit.fit()
    term.focus()
    return term

    function LoadLigaturesAddon() {
        try {
            const newLocal = new LigaturesAddon.LigaturesAddon();
            term.loadAddon(newLocal);
        } catch (error) {
            console.error(error);
        }
    }
}
socket_int = (term, app) => {
    let localhost = window.location.host
    var socket = new WebSocket('ws://' + localhost + '/ws');
    const sendTextData = (data) => {
        if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(data));
            // console.log('Sent to server:', data);
        } else {
            console.error('WebSocket connection is not open.');
        }
    }
    ws_sendTextData = sendTextData
    const resizecall = () => {
        let call = "resize"
        let rows = term.rows, cols = term.cols
        sendTextData({ call, cols, rows })
    }
    socket.onopen = function (event) {
        console.log("Connection opened");
        call = "init"
        let rows = term.rows, cols = term.cols
        let host = window.location.host
        sendTextData({ call, cols, rows, host })
    };


    socket.binaryType = "blob";
    socket.onmessage = function incoming(evt) {
        const handleMessage = (data) => {
            // 处理解码后的数据
            var { Call, Output } = data
            if (Call == "term") {
                term.write(Output)
            }
            else if (Call == "zoom") {
                var { Zoom } = data
                let fontsize = get_font_size()
                if (Zoom) {
                    fontsize++
                } else {
                    fontsize--
                }
                set_font_size(fontsize)
                window.location.reload()
                console.log("zoom", Zoom)
            } else if (Call == "openfile") {
                console.log("openfile",
                    data.Filename)
                let ext = getFileExtension(data.Filename)
                if (is_image(ext)) {
                    app.popimage(data.Filename)
                } else if (is_md(ext)) {
                    app.popmd(data.Filename)
                }
            } else if (call_on_copy == Call) {
                var copybtn = document.getElementById("copybtn")
                let text = data.SelectedString
                copybtn.addEventListener("click", () => {
                    navigator.clipboard.writeText(text)
                        .then(() => {
                            console.log('文本已复制到剪贴板');
                            alert('文本已复制到剪贴板');
                        })
                        .catch((err) => {
                            console.error('无法复制文本: ', err);
                            alert('无法复制文本，请检查是否支持此功能');
                        });
                })
                console.log(data.SelectedString)
            }
            // console.log("Received: ", event.data);
        }
        try {
            var reader = new FileReader();
            reader.readAsArrayBuffer(evt.data);
            reader.addEventListener("loadend", function (e) {
                const buffer = new Uint8Array(e.target.result);  // arraybuffer object
                const message = msgpack5().decode(buffer);
                handleMessage(message)
            });
        } catch (error) {
            console.error('Failed to decode data:', error);
        }
    };
    socket.onclose = function (event) {
        console.error("Connection closed");
    };
    window.addEventListener('resize', function (evt) {
        let f = new fullscreen_check(term)
        f.resize()
    })
    term.onResize((size) => {
        console.log("event resize", size)
        resizecall()
    })
}
main = () => {
    var app = app_init()
    var term = term_init(app)
    socket_int(term, app)
}
main()
function set_font_size(fontSize) {
    window.localStorage.setItem("fontsize", fontSize);
}
function get_font_size() {
    var fontSize = window.localStorage.getItem("fontsize");
    if (fontSize == undefined || fontSize == "undefined") {
        fontSize = 12;
    }
    return fontSize;
}
