package web_test

import (
	// "os"
	"fmt"
	"os"
	"testing"

	web "zen108.com/lspvi/pkg/ui/xterm"
)

func TestXxx(t *testing.T) {
	buf, _ := os.ReadFile("/home/z/dev/lsp/goui/README.md")
	// web.(buf)
	if b, e := web.ChangeLink(buf, false, "/md/"); e == nil {
		x := string(b)
		fmt.Println(x)
	}
}
