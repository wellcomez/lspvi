package mainui

import (
	"fmt"
	"log"
	"net/http"
)

func servmain(root string, port int) {
	// 指定要浏览的目录
	// dirToServe := "./public"

	// 使用 http.Dir 来指定目录
	fileSystem := http.Dir(root)

	// 使用 http.FileServer 来提供静态文件服务
	fs := http.FileServer(fileSystem)

	// 设置路由处理器
	http.Handle("/", fs)

	// 启动服务器
	log.Printf("Starting server at port %d",port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d",port), nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
