package web

import (
	// "fmt"
	"bytes"
	"path/filepath"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"zen108.com/lspvi/pkg/debug"
)

type RegexpLinkTransformer struct {
	root string
}

func (t *RegexpLinkTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// source := reader.Source()

	// Walk the AST in depth-first fashion and apply transformations
	err := ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		// Each node will be visited twice, once when it is first encountered (entering), and again
		// after all the node's children have been visited (if any). Skip the latter.
		if !entering {
			return ast.WalkContinue, nil
		}
		// Skip the children of existing links to prevent double-transformation.
		if node.Kind() == ast.KindImage {
			textNode := node.(*ast.Image)
			path := string(textNode.Destination)
			if !filepath.IsAbs(path) {
				if !strings.HasPrefix(path, "http") {
					path = filepath.Join(t.root, path)
					textNode.Destination = []byte(path)
				}
			}
			return ast.WalkSkipChildren, nil
		}
		if node.Kind() == ast.KindLink || node.Kind() == ast.KindAutoLink {
			textNode := node.(*ast.Link)
			path := string(textNode.Destination)
			if !filepath.IsAbs(path) {
				if !strings.HasPrefix(path, "http") {
					path = filepath.Join(t.root, path)
					textNode.Destination = []byte(path)
				}
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		debug.ErrorLog("md", "Error encountered while transforming AST:", err)
	}
}

type CSSInserter struct {
	css string
}

// Extend implements the goldmark.Extender interface
func (c *CSSInserter) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&htmlRendererWithCSS{c.css}, 100),
		),
	)
}

// htmlRendererWithCSS is a custom HTML renderer that inserts CSS
type htmlRendererWithCSS struct {
	css string
}

// RegisterFuncs implements the renderer.NodeRenderer interface
func (r *htmlRendererWithCSS) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Register the default HTML renderer functions
	html.NewRenderer().RegisterFuncs(reg)

	// Add a function to insert the CSS into the head
	reg.Register(ast.KindDocument, r.insertCSS)
}

// insertCSS inserts the CSS into the document head
func (r *htmlRendererWithCSS) insertCSS(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// Write the opening <head> tag
		_, _ = w.WriteString("<head>")
		sss := []string{
			`<link rel="stylesheet" href="/static/highlightjs/styles/default.min.css">`,
			`<script src="/static/highlightjs/highlight.min.js"></script>`,
			`<script src="/static/highlightjs/languages/go.min.js"></script>`,
			`<script >hljs.highlightAll();</script>`,
		}
		_, _ = w.WriteString(strings.Join(sss, "\n"))
		// Write the CSS
		_, _ = w.WriteString("<style>")
		_, _ = w.WriteString(r.css)
		_, _ = w.WriteString("</style>")
	} else {
		// Write the closing </head> tag
		_, _ = w.WriteString("</head>")
	}
	return ast.WalkContinue, nil
}

var css = "img { max-width: 100%;width: 90% }"

func ChangeLink(source []byte, root string) (ret []byte, err error) {
	// LinkPattern: regexp.MustCompile(`TICKET-\d+`),
	// ReplUrl:     []byte("https://example.com/TICKET?query=$0"),
	// Goldmark supports multiple AST transformers and runs them sequentially in order of priority.
	// Setup goldmark with the markdown renderer and our transformer
	link_parser := create_link_parser(root)
	x := css_extentsion()
	gm := goldmark.New(
		goldmark.WithRenderer(markdown.NewRenderer()),
		goldmark.WithParserOptions(link_parser),
		goldmark.WithExtensions(x),
	)
	// source := []byte(`[Link](./old-path.md)`)
	reader := text.NewReader(source)
	document := gm.Parser().Parse(reader)
	var buf bytes.Buffer
	err = gm.Renderer().Render(&buf, source, document)
	if err == nil {
		ret = buf.Bytes()
	}
	return
}

func css_extentsion() *CSSInserter {
	x := &CSSInserter{
		css: css,
	}
	return x
}

func create_link_parser(root string) parser.Option {
	transformer := RegexpLinkTransformer{
		root: root,
	}
	prioritizedTransformer := util.Prioritized(&transformer, 0)
	x := parser.WithASTTransformers(prioritizedTransformer)
	return x
}
