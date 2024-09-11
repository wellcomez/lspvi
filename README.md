## About lspvi

lspvi is tui code reader powsered by lsp. **Just code reader not editor**

It provides goto define ,goto declaration, call hierarchy,reference  etc which also provided by common ide.

But compared with vscode,neovim ..., lspvi provide  more powerful function for call hierarchy,reference  to analyze code.

It can prvoides **UML call sequence** which can help you to understand the code. 

It also support fzf search for files and text, symbol ,call reference,live grep, word grep. 

All above features are built in single binary not need to install plugin. Binary also includes built-in web server by which you can view uml in browser.

![terminal](screen1.png)

![uml](main.png)

## Mouse Support

Although it run in terminal, it supports mouse click, wheel scrolling as well as gui. So if you are not vimer, it is also easy to use.


## User in Browser/Remotely
- start as web server
~~~sh
lspvi -gui
~~~
- open browser 

start with 13000
[http:://localhost:1300x](http:://localhost:13000)

![webpng](web.png)
- with power of xterm.js 

## lsp supported

- gopls
- clangd
- python
- typescript /javascript

**prerequisite**:

above lsp serveries you have installed

## Install

```sh
git submodule init  && git submodule update --recursive
pushd pkg/lspr
bash build.sh -w
popd
go build
./lspvi --root "prj root"
```

## platform /os

- MACOS 
- linux 
- windows

## docker

```sh
cd docker
```

1. build docker
   
   ```sh
   ./dockbuild.sh create
   ```
2. run docker 
   
   ```sh
   ./dockbuild.sh 
   ```

# Keymap

It supports most of common keymap of [neo]vim. **Space** will invoke menu and  choose **help** or   **space + h** will invoke keymap help.

![keymap](keymap.png)
| key            | function               |
| -------------- | ---------------------- |
| escape + f     | file in file           |
| escape + *     | file in file vi        |
| escape + /     | search mode            |
| escape + 0     | goto line head         |
| escape + c     | goto callin            |
| escape + r     | goto refer             |
| escape + k     | up                     |
| escape + %     | match                  |
| escape + Up    | up                     |
| escape + h     | left                   |
| escape + Left  | left                   |
| escape + l     | right                  |
| escape + Right | right                  |
| escape + j     | down                   |
| escape + Down  | down                   |
| escape + e     | word right             |
| escape + b     | word left              |
| escape + yy    | Copy                   |
| escape + y     | Copy                   |
| escape + gd    | goto define            |
| escape + gr    | goto refer             |
| escape + gg    | goto first line        |
| escape + G     | goto first line        |
| escape + B     | Bookmark               |
| escape + xf    | goto file explorer     |
| space + f      | picker file            |
| space + fw     | grep word              |
| space + ws     | query workspace symbol |
| space + hq     | quickfix history       |
| space + r      | reference              |
| space + hh     | history                |
| space + o      | open symbol            |
| menu + o       | open symbol            |
| menu + q       | quickfix history       |
| menu + r       | reference              |
| menu + B       | bookmark               |
| menu + g       | live grep              |
| menu + hh      | history                |
| menu +         | colorscheme            |
| menu + wk      | workspace              |
| menu + fw      | grep word              |
| menu + ff      | Search in files        |
| menu + f       | picker file            |
| menu + h       | help                   |
| menu + ws      | query workspace symbol |
| menu + Q       | Quit                   |
| ctrl+w j       | next window down       |
| ctrl+w k       | next window up         |
| ctrl+w h       | next window left       |
| ctrl+w l       | next window right      |
| ctrl+w Down    | next window down       |
| ctrl+w Up      | next window up         |
| ctrl+w Left    | next window left       |
| ctrl+w Right   | next window right      |
| Ctrl-P         | picker file            |
| Tab            | tab                    |