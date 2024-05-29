package middleware

import (
	"dogker/lintang/container-service/biz/mw/jwt"

	"github.com/cloudwego/hertz/pkg/app"
)

func Protected() []app.HandlerFunc {
	mwJwt := jwt.GetJwtMiddleware()
	mwJwt.MiddlewareInit()
	return []app.HandlerFunc{
		mwJwt.MiddlewareFunc(),
	}

}
