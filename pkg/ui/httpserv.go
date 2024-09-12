//go:build ignore

package mainui

import (
	httpsrv "zen108.com/lspsrv/pkg"
)

func servmain(root string, port int, cb func(port int)) {
	httpsrv.StartServer(root, port, cb)
}
