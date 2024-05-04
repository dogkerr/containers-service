// Code generated by hertz generator.

package hello

import (
	"dogker/lintang/container-service/biz/mw/jwt"

	"github.com/cloudwego/hertz/pkg/app"
)

func rootMw() []app.HandlerFunc {
	// your code...
	return nil
}

func _method1Mw() []app.HandlerFunc {
	// your code...
	return nil
}

func _apiMw() []app.HandlerFunc {
	// your code...
	return nil
}

func _v1Mw() []app.HandlerFunc {
	// your code...
	return nil
}

func _helloMw() []app.HandlerFunc {
	// your code...
	mwJwt := jwt.GetJwtMiddleware()
	mwJwt.MiddlewareInit()
	return []app.HandlerFunc{
		mwJwt.MiddlewareFunc(),
	}
}
