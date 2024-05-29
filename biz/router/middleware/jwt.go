package middleware

import (
	"context"
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

func Cors() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Request.Header.Set("Access-Control-Allow-Origin", "*")
		c.Request.Header.Set("Access-Control-Allow-Credentials", "true")
		c.Request.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Request.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if string(c.GetRequest().Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next(ctx)
	}
}
