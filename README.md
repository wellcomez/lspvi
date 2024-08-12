## About lspvi
lspvi is tui code viewer powser by lsp.
It provides godefine ,goto declaration, call hierarchy,reference  etc.
Compared with vscode,neovim ..., lspvi provides  more powser feature in call hierarchy,reference etc. It can prvoides UML call sequence which can help you to understand the code.

![terminal](screen1.png)

![uml](main.png)
## Mouse Support
Although it runs under terminal, but it supports mouse click, wheel scroll.
## Keymap 
It supports most of common keymap of nvim. "Space" will invoke menu and  choose "help" /" space + h" will invoke help.

## Code/Uml  view in browser
Because of the limitation of terminal, it can just disp text call sequence instead of png. With help of browser, you can view the call sequence in browser.

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
Macos/linux I have tested

