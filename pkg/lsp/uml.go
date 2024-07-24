package lspcore

import (
	"fmt"
	"log"
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
func (entry CallStackEntry) uml_class_name() string {
	if entry.PtrSymobl == nil {
		return ""
	}
	return strings.ReplaceAll(entry.PtrSymobl.classname, ":", ".")
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
			rightPrefix = s.uml_class_name() + "::"
		}
		right := rightPrefix + s.symboldefine_name()
		if strings.Index(right, "ProcessInternal") > 0 {
			log.Println(right, s.symboldefine_name(), s.uml_class_name())
		}
		if len(ret) == 0 {
			title = fmt.Sprintf("==%s==", right)
		}
		if !s.isFunction() {
			left := s.uml_class_name()
			if caller != nil {
				if caller.isFunction() {
					left = caller.symboldefine_name()
				} else {
					if caller.uml_class_name() != s.uml_class_name() {
						left = caller.uml_class_name()
					}
				}
			}
			ret = append(ret, fmt.Sprintf("%s -> %s", left, right))
		} else {
			if caller != nil {
				left := caller.uml_class_name()
				if caller.isFunction() {
					left = caller.symboldefine_name()
				}
				ret = append(ret, fmt.Sprintf("%s -> %s", left, right))
			}
		}
		caller = s
	}
	markBegin := ""
	if markdown {
		markBegin = "```plantuml"
	}
	var black = "skinparam monochrome true"

	sss := []string{markBegin, "@startuml", black, "autoactivate on", title}
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
