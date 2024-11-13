package web

import (
	// "fmt"
	"bytes"
	"path/filepath"
	"regexp"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"zen108.com/lspvi/pkg/debug"
)

type RegexpLinkTransformer struct {
	LinkPattern *regexp.Regexp
	ReplUrl     []byte
	root        string
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
func (t *RegexpLinkTransformer) LinkifyText(node *ast.Text, source []byte) {
	parent := node.Parent()
	tSegment := node.Segment
	match := t.LinkPattern.FindIndex(tSegment.Value(source))
	if match == nil {
		return
	}
	// Create a text.Segment for the link text.
	lSegment := text.NewSegment(tSegment.Start+match[0], tSegment.Start+match[1])

	// Insert node for any text before the link
	if lSegment.Start != tSegment.Start {
		bText := ast.NewTextSegment(tSegment.WithStop(lSegment.Start))
		parent.InsertBefore(parent, node, bText)
	}

	// Insert Link node
	link := ast.NewLink()
	link.AppendChild(link, ast.NewTextSegment(lSegment))
	link.Destination = t.LinkPattern.ReplaceAll(lSegment.Value(source), t.ReplUrl)
	parent.InsertBefore(parent, node, link)

	// Update original node to represent the text after the link (may be empty)
	node.Segment = tSegment.WithStart(lSegment.Stop)

	// Linkify remaining text if not empty
	if node.Segment.Len() > 0 {
		t.LinkifyText(node, source)
	}
}

// LinkPathReplacer is a custom extension to replace link paths
type LinkPathReplacer struct {
	newRootPath string
}

// Extend implements the goldmark.Extender interface
func (l *LinkPathReplacer) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(l, 100)))
}

// RenderNode modifies the link paths during rendering
func (l *LinkPathReplacer) RenderNode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if link, ok := node.(*ast.Link); ok {
			// Replace the link destination
			link.Destination = []byte(strings.ReplaceAll(string(link.Destination), "./", l.newRootPath))
			link.Destination = []byte(strings.ReplaceAll(string(link.Destination), "../", l.newRootPath))
		}
	}
	return ast.WalkContinue, nil
}
func ChangeLink(source []byte, root string) (ret []byte, err error) {
	// LinkPattern: regexp.MustCompile(`TICKET-\d+`),
	// ReplUrl:     []byte("https://example.com/TICKET?query=$0"),
	// Goldmark supports multiple AST transformers and runs them sequentially in order of priority.
	// Setup goldmark with the markdown renderer and our transformer
	link_parser := create_link_parser(root)
	gm := goldmark.New(
		goldmark.WithRenderer(markdown.NewRenderer()),
		goldmark.WithParserOptions(link_parser),
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

func create_link_parser(root string) parser.Option {
	transformer := RegexpLinkTransformer{
		root: root,
	}
	prioritizedTransformer := util.Prioritized(&transformer, 0)
	x := parser.WithASTTransformers(prioritizedTransformer)
	return x
}
func Replace(source []byte) (ret []byte, err error) {
	// Create a new Goldmark instance with the custom extension
	md := goldmark.New(
		goldmark.WithExtensions(&LinkPathReplacer{newRootPath: "/new/path/"}),
	)
	// renderer := glo.NewRenderer(.WithHeadingStyle(markdown.HeadingStyleATX))
	// Parse and render the Markdown content
	reader := text.NewReader(source)
	document := md.Parser().Parse(reader)
	var buf bytes.Buffer
	err = md.Renderer().Render(&buf, source, document)
	if err == nil {
		ret = buf.Bytes()
	}
	return
}
