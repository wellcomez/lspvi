package mainui

import (
	"fmt"
	"strings"

	lspcore "zen108.com/lspvi/pkg/lsp"
)



func caller_to_listitem(caller *lspcore.CallStackEntry, root string) string {
	if caller == nil {
		return ""
	}
	callerstr := fmt.Sprintf(" [%s %s:%d]", caller.Name,
		strings.TrimPrefix(
			caller.Item.URI.AsPath().String(), root),
		caller.Item.Range.Start.Line)
	return callerstr
}
