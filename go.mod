module zen108.com/lspvi

go 1.23.0

toolchain go1.23.2

//replace zen108.com/lspsrv => ./pkg/lspr

replace github.com/pgavlin/femto => ./pkg/femto

replace github.com/smacker/go-tree-sitter => ./pkg/go-tree-sitter

replace github.com/tree-sitter/tree-sitter-markdown => ./pkg/tree-sitter-markdown

// replace github.com/tree-sitter-grammars/tree-sitter-yaml/bindings/go => ./pkg/tree-sitter-yaml/bindings/go
replace github.com/tree-sitter-grammars/tree-sitter-yaml => ./pkg/tree-sitter-yaml/bindings/go

replace github.com/tree-sitter-grammars/tree-sitter-toml => ./pkg/tree-sitter-toml/bindings/go

replace github.com/creack/pty => ./pkg/photostorm/pty

replace github.com/iyzyi/aiopty => ./pkg/aiopty

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
	github.com/tree-sitter/go-tree-sitter v0.24.0
	github.com/tree-sitter/tree-sitter-go v0.23.1
	github.com/tree-sitter/tree-sitter-java v0.23.2
	github.com/tree-sitter/tree-sitter-json v0.24.1
	github.com/tree-sitter/tree-sitter-python v0.23.2

	// github.com/tree-sitter/tree-sitter-markdown v0.3.2
	// github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
	// github.com/tectiv3/go-lsp v0.0.0-20240419022041-0a0a5672827e
	github.com/vmihailenco/msgpack/v5 v5.4.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/bmatcuk/doublestar v1.3.4
	github.com/boyter/go-string v1.0.5
	github.com/iyzyi/aiopty v0.1.0
	github.com/sergi/go-diff v1.3.1
	github.com/tectiv3/go-lsp v0.0.0-20240419022041-0a0a5672827e
	github.com/tree-sitter-grammars/tree-sitter-lua v0.2.0
	github.com/tree-sitter-grammars/tree-sitter-toml v0.0.0-00010101000000-000000000000
	github.com/tree-sitter/tree-sitter-bash v0.23.1
	github.com/tree-sitter/tree-sitter-c v0.21.5-0.20240818205408-927da1f210eb
	github.com/tree-sitter/tree-sitter-cpp v0.22.4-0.20240818224355-b1a4e2b25148
	github.com/tree-sitter/tree-sitter-css v0.23.0
	github.com/tree-sitter/tree-sitter-html v0.20.5-0.20240818004741-d11201a263d0
	github.com/tree-sitter/tree-sitter-javascript v0.21.5-0.20240818005344-15887341e5b5
	github.com/tree-sitter/tree-sitter-markdown v0.0.0-00010101000000-000000000000
	github.com/tree-sitter/tree-sitter-ruby v0.21.1-0.20240818211811-7dbc1e2d0e2d
	github.com/tree-sitter/tree-sitter-rust v0.21.3-0.20240818005432-2b43eafe6447
	github.com/tree-sitter/tree-sitter-typescript v0.23.0
	github.com/yuin/goldmark v1.5.4
)

require (
	cloud.google.com/go v0.104.0 // indirect
	cloud.google.com/go/compute v1.10.0 // indirect
	cloud.google.com/go/iam v0.5.0 // indirect
	cloud.google.com/go/storage v1.27.0 // indirect
	github.com/alecthomas/chroma/v2 v2.2.0 // indirect
	github.com/aws/aws-sdk-go v1.44.122 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.6.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82 // indirect
	github.com/teekennedy/goldmark-markdown v0.3.0 // indirect
	github.com/tree-sitter-grammars/tree-sitter-yaml v0.6.1 // indirect
	github.com/tree-sitter-grammars/tree-sitter-zig v1.0.2 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.abhg.dev/goldmark/toc v0.10.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.100.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221025140454-527a21cfbd71 // indirect
	google.golang.org/grpc v1.50.1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

require (
	github.com/arduino/go-paths-helper v1.12.1 // indirect
	github.com/cavaliergopher/grab/v3 v3.0.1
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/gdamore/tcell/v2 v2.7.4
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pgavlin/femto v0.0.0-20201224065653-0c9d20f9cac4
	github.com/rivo/tview v0.0.0-20240807205129-e4c497cc59ed
	github.com/rivo/uniseg v0.4.7 // indirect
	go.bug.st/json v1.15.6
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/term v0.23.0
	golang.org/x/text v0.17.0 // indirect
)
