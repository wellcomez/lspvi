module zen108.com/lspvi

go 1.21.3

//replace zen108.com/lspsrv => ./pkg/lspr

replace github.com/pgavlin/femto => ./pkg/femto

replace github.com/smacker/go-tree-sitter => ./pkg/go-tree-sitter

require (
	github.com/charlievieth/fastwalk v1.0.8
	github.com/reinhrst/fzf-lib v0.9.0
	github.com/sourcegraph/jsonrpc2 v0.2.0
// github.com/tectiv3/go-lsp v0.0.0-20240419022041-0a0a5672827e
// zen108.com/lspsrv v0.0.0-00010101000000-000000000000
)

require (
	// fyne.io/fyne/v2 v2.5.1
	github.com/akiyosi/qt v0.0.0-20240304155940-b43fff373ad5
	github.com/atotto/clipboard v0.1.4
	github.com/creack/pty v1.1.23
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.4.1
	// github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
	// github.com/tectiv3/go-lsp v0.0.0-20240419022041-0a0a5672827e
	github.com/vmihailenco/msgpack/v5 v5.4.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
	github.com/tectiv3/go-lsp v0.0.0-20240419022041-0a0a5672827e
)

require (
	github.com/boyter/go-string v1.0.5 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)

require (
	github.com/arduino/go-paths-helper v1.12.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/gdamore/tcell/v2 v2.7.4
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pgavlin/femto v0.0.0-20201224065653-0c9d20f9cac4
	github.com/rivo/tview v0.0.0-20240807205129-e4c497cc59ed
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570
	go.bug.st/json v1.15.6
	golang.org/x/sys v0.24.0
	golang.org/x/term v0.23.0
	golang.org/x/text v0.17.0 // indirect
)
