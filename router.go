// Code generated by hertz generator.

package main

import (
	handler "dogker/lintang/container-service/biz/handler"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)

	// your code ...
}
