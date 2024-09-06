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
let rows = 50
let cols = 80
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
app_init = () => {
    let app = new Vue({
        el: '#app',
        data: {
            message: 'Hello Vue!',
            isVisible: false,
            imageurl: "",
        },
        methods: {
            onhide() {
                this.isVisible = false
            },
            on_wheel(evt) {
                return this.isVisible == true
            },
            popimage(image) {
                this.isVisible = true
                this.imageurl = image
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
    var term = new Terminal({
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
    let f = new fullscreen_check(term)
    f.resize(false)
    fit.fit()
    term.focus()
    return term
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
            else if (Call == "openfile") {
                console.log("openfile",
                    data.Filename)
                let ext = getFileExtension(data.Filename)
                if (is_image(ext)) {
                    app.popimage(data.Filename)
                }
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