// import Vue from "vue";
import { Term } from "./term"
import msgpack5 from "msgpack5";
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
const app_init = () => {
    // let md = md_init()
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
                    // md.on_wheel(evt)
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
                // md.render(u)
            }
        }
    })

    let wheel = (evt) => {
        console.log(evt)
    }
    document.addEventListener("wheel", wheel)
    const div = document.getElementById('terminal')
    if (div) {
        div.addEventListener("wheel", wheel)
    }
    return app
}
class RemoteConn {
    constructor(socket) {
        this.socket = socket
    }
}
var ws_sendTextData
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
    let termobj = new Term()
    let app = app_init()
    termobj.setapp(app)
    socket_int(termobj, app, (cmdline) => {
        termobj.start_xterm(true, cmdline)
    })

}
export default { main } 