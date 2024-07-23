package lspcore

import (
	"fmt"
	"strings"

	"github.com/tectiv3/go-lsp"
)

// 假设以下类型和函数已经被定义：
// type WorkSpaceSymbol struct{}
//
//	type CallNode struct {
//	    symboldefine *Symbol
//	    // 其他字段...
//	}
//
//	type Symbol struct {
//	    kind SymbolKind
//	    name string
//	    // 其他字段...
//	}
//
// type SymbolKind int
// const (
//
//	Function SymbolKind = iota
//	// 其他符号种类...
//
// )
// func (c *CallNode) get_cls() *Symbol { /* 实现 */ }
// func (w *WorkSpaceSymbol) find(c *CallNode) { /* 实现 */ }
func (entry CallStackEntry) isFunction() bool {
	return entry.Item.Kind == lsp.SymbolKindFunction
}
func (entry CallStackEntry) class_name() string {
	if entry.PtrSymobl == nil {
		return ""
	}
	return entry.PtrSymobl.classname
}
func (entry CallStackEntry) symboldefine_name() string {
	return entry.Item.Name
}

func (call CallStack) Uml(markdown bool) string {
	ret := make([]string, 0)
	var caller *CallStackEntry = nil

	title := ""
	for _, s := range call.Items {
		rightPrefix := ""
		if !s.isFunction() {
			rightPrefix = strings.ReplaceAll(s.class_name(), "::", ".")+"::"
		}
		right := rightPrefix + s.symboldefine_name()
		if len(ret) == 0 {
			title = fmt.Sprintf("==%s==", right)
		}
		if !s.isFunction() {
			left := s.class_name()
			if caller != nil {
				if caller.isFunction() {
					left = caller.symboldefine_name()
				} else {
					if caller.class_name() != s.class_name() {
						left = caller.class_name()
					}
				}
			}
			ret = append(ret, fmt.Sprintf("%s -> %s", strings.Replace(left, "::", ".", -1), right))
		} else {
			if caller != nil {
				left := caller.class_name()
				if !caller.isFunction() {
					left = caller.symboldefine_name()
				}
				ret = append(ret, fmt.Sprintf("%s -> %s", strings.Replace(left, "::", ".", -1), right))
			}
		}
		caller = s
	}
	markBegin := ""
	if markdown {
		markBegin = "```plantuml"
	}
	sss := []string{markBegin, "@startuml", "autoactivate on", title}
	sss = append(sss, ret...)
	markEnd := ""
	if markdown {

		markEnd = "```\n\n\n"
	}
	sss = append(sss, "@enduml", markEnd)

	return strings.Join(sss, "\n")
}

// 假设applyFix函数被定义为应用fix函数到stack中的每个元素
// func applyFix(stack []*CallNode, fixFunc func(*CallNode) *CallNode) []*CallNode {
// 	for i, node := range stack {
// 		stack[i] = fixFunc(node)
// 	}
// 	return stack
// }
