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
	return strings.ReplaceAll(entry.PtrSymobl.Classname, ":", ".")
}
func (entry CallStackEntry) symboldefine_name() string {
	return entry.Item.Name
}
func (call CallStack) newuml() (ret []string, title string) {
	pre := []string{"autonumber","actor 0 #red"}
	ret = []string{}
	added := make(map[string]bool)
	actor := "0"
	for _, s := range call.Items {
		right := ""
		current_actor := ""
		if current_actor = s.uml_class_name(); len(current_actor) == 0 {
			current_actor = s.Item.URI.AsPath().Base()
			if _, ok := added[current_actor]; !ok {
				s := fmt.Sprintf("collections  %s", current_actor)
				pre = append(pre, s)
				added[current_actor] = true
			}

		}
		right = fmt.Sprint(current_actor, ":", s.symboldefine_name())

		if len(ret) == 0 {
			title = fmt.Sprintf("==%s==", right)
		}
		ret = append(ret, fmt.Sprintf("%s -> %s", actor, right))
		actor = current_actor
	}
	return append(pre, ret...), title
}

func (call CallStack) Uml(markdown bool) string {
	//go use function as method because treate package as object
	ret, title := call.newuml()
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

func (call CallStack) olduml() ([]string, string) {
	ret := make([]string, 0)
	var caller *CallStackEntry = nil

	title := ""
	for _, s := range call.Items {
		rightPrefix := ""
		if !s.isFunction() {

			if len(s.uml_class_name()) > 0 {
				rightPrefix = s.uml_class_name() + "::"
			}
		}
		right := rightPrefix + s.symboldefine_name()
		if strings.Index(right, "ProcessInternal") > 0 {
			log.Println(right, s.symboldefine_name(), s.uml_class_name())
		}
		if len(ret) == 0 {
			title = fmt.Sprintf("==%s==", right)
		}
		if caller != nil {
			if !s.isFunction() {
				left := s.uml_class_name()
				if caller.isFunction() || len(left) == 0 {
					left = caller.symboldefine_name()
				} else {
					if len(caller.uml_class_name()) > 0 && caller.uml_class_name() != s.uml_class_name() {
						left = caller.uml_class_name()
					} else {
						left = caller.symboldefine_name()
					}
				}
				ret = append(ret, fmt.Sprintf("%s -> %s", left, right))
			} else {
				left := caller.uml_class_name()
				if caller.isFunction() || len(left) == 0 {
					left = caller.symboldefine_name()
				}
				ret = append(ret, fmt.Sprintf("%s -> %s", left, right))
			}
		}
		caller = s
	}
	return ret, title
}

// 假设applyFix函数被定义为应用fix函数到stack中的每个元素
// func applyFix(stack []*CallNode, fixFunc func(*CallNode) *CallNode) []*CallNode {
// 	for i, node := range stack {
// 		stack[i] = fixFunc(node)
// 	}
// 	return stack
// }
