package lspcore

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tectiv3/go-lsp"
)

type Body struct {
	Subline  []string
	Location lsp.Location
}

// 定义一个结构体来表示行和字符的位置

// SubLine 函数用于从多行文本中提取子集
func SubLine(begin, end lsp.Position, lines []string) ([]string, error) {
	if begin.Line > len(lines) {
		return nil, fmt.Errorf("begin line out of range")
	}
	if end.Line > len(lines) {
		return nil, fmt.Errorf("end line out of range")
	}
	subline := lines[begin.Line : end.Line+1]
	e := end.Character
	if begin.Character >= len(subline[0]) {
		return subline, nil
	}
	if e >= len(subline[0]) {
		return subline, nil
	}
	if begin.Line == end.Line {
		if e < 0 {
			e = -1
		}

		line := subline[0]
		if begin.Character < len(line) && e < len(line) {
			subline[0] = line[begin.Character:e]
		}
	} else {
		subline[0] = subline[0][begin.Character:]
		if e < 0 {
			e = -1
		}
		line := subline[len(subline)-1]
		if len(line) > e {
			line = line[:e]
		}
		subline[len(subline)-1] = line
	}

	return subline, nil
}

var use_uri = []string{".txt", ".json"} // 示例值，根据实际情况调整
func getext(path string) string {
	return filepath.Ext(path)
}
func from_uri(path string) string {
	// 实现取决于具体的 URI 处理逻辑
	return path // 返回原始路径作为示例
}

// from_file 函数用于处理文件路径或 URI
func from_file(path string) string {
	ext := getext(path)
	for _, uriExt := range use_uri {
		if ext == uriExt {
			return from_uri(path)
		}
	}

	// 处理 file:// 或 file: 前缀
	return strings.ReplaceAll(strings.ReplaceAll(path, "file://", ""), "file:", "")
}

func NewBody(location lsp.Location) (*Body, error) {
	_range := location.Range
	begin := _range.Start
	end := _range.End

	// 读取文件内容
	content, err := ioutil.ReadFile(from_file(location.URI.AsPath().String()))
	if err != nil {
		return nil, err
	}

	// 将内容分割成行
	lines := strings.Split(string(content), "\n")

	// 提取子行
	subline, err := SubLine(begin, end, lines)
	if err == nil {
		return &Body{
			Subline:  subline,
			Location: location,
		}, nil
	}
	return nil, err
}

// String 方法返回 Body 的字符串表示
func (b *Body) Info() string {
	return fmt.Sprintf("%s %d:%d %d:%d", b.String(), b.Location.Range.Start.Line, b.Location.Range.Start.Character, b.Location.Range.End.Line, b.Location.Range.End.Character)
}
func (b *Body) String() string {
	return strings.Join(b.Subline, "\n")
}

func main() {
	// 示例用法
	lines := []string{"This is line one", "This is line two", "This is line three"}
	begin := lsp.Position{0, 5}
	end := lsp.Position{1, 10}
	if sublines, err := SubLine(begin, end, lines); err == nil {
		for _, line := range sublines {
			fmt.Println(line)
		}
	}
}
