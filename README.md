## About lspvi

lspvi is tui code reader powsered by lsp. **Just code reader not editor**

It provides goto define ,goto declaration, call hierarchy,reference  etc which also provided by common ide.

But compared with vscode,neovim ..., lspvi provide  more powerful function for call hierarchy,reference etc.

It can prvoides **UML call sequence** which can help you to understand the code. 

It also support fzf search for files and text, symbol ,call reference,live grep, word grep. 

All above features are built in single binary not need to install plugin. Binary also includes built-in web server by which you can view uml in browser.

![terminal](screen1.png)

![uml](main.png)
## Mouse Support
Although it run in terminal, it supports mouse click, wheel scrolling as well as gui. So if you are not vimer, it is also easy to use.

## Keymap 
It supports most of common keymap of [neo]vim. **Space** will invoke menu and  choose **help** or   **space + h** will invoke keymap help.

![keymap](keymap.png)

## Code/Uml  view in browser
Because of the limitation of terminal, under which  you can still view text call sequence. If you want to view png clearly, , you can open browser and see the call sequence in browser.

![web](web.png)

## lsp supported
- gopls
- clangd
- python

## Install
~~~sh
git submodule init  && git submodule update --recursive
pushd pkg/lspr
bash build.sh -w
popd
go build
./lspvi --root "prj root"
~~~

## platform /os 
- MACOS 
- linux 

