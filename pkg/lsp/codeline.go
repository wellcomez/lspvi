package lspcore

import (
	"fmt"

	"github.com/tectiv3/go-lsp"
)

// 定义一个结构体来表示行和字符的位置

// SubLine 函数用于从多行文本中提取子集
func SubLine(begin, end lsp.Position, lines []string) []string {
	subline := lines[begin.Line : end.Line+1]

	if begin.Line == end.Line {
		e := end.Character + 1
		if e < 0 {
			e = -1
		}
		subline[0] = subline[0][begin.Character:e]
	} else {
		subline[0] = subline[0][begin.Character:]
		e := end.Character + 1
		if e < 0 {
			e = -1
		}
		subline[len(subline)-1] = subline[len(subline)-1][:e]
	}

	return subline
}

func main() {
	// 示例用法
	lines := []string{"This is line one", "This is line two", "This is line three"}
	begin := lsp.Position{0, 5}
	end := lsp.Position{1, 10}
	sublines := SubLine(begin, end, lines)
	for _, line := range sublines {
		fmt.Println(line)
	}
}
