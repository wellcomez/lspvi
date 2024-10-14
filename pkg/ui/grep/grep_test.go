package grep

import (
	"os"
	"testing"

	str "github.com/boyter/go-string"
)
func TestGoString(t *testing.T) {
	if _,err:=os.ReadFile("/home/z/dev/lsp/goui/pkg/ui/gogrepimpl.go");err==nil{
		// sss:=string(d)
		ret:=str.IndexAll("err123456\nerr","err",-1)	
		for _, v := range ret {
			t.Log(v)
		}
	}

}