package middleware

import (
	"context"
	"dogker/lintang/container-service/biz/mw/jwt"

	"github.com/cloudwego/hertz/pkg/app"
)

func Protected() []app.HandlerFunc {
	mwJwt := jwt.GetJwtMiddleware()
	mwJwt.MiddlewareInit()
	// fun := func(ctx context.Context, c *app.RequestContext) {
	// 	origin := c.Request.Header.Get("Origin")
	// 	if origin != "" {
	// 		c.Response.Header.Set("Access-Control-Allow-Origin", origin)
	// 		c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	// 		c.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, authorization, content-type, accept, origin, Cache-Control, X-Requested-With")
	// 		c.Response.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
	// 	}
	// 	// tadi preflightnya belum ada


	// }//gakbisa
	return []app.HandlerFunc{
		mwJwt.MiddlewareFunc(),
	}


}

func Cors() app.HandlerFunc {

	return func(ctx context.Context, c *app.RequestContext) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Response.Header.Set("Access-Control-Allow-Origin", origin)
			c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			c.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, authorization, content-type, accept, origin, Cache-Control, X-Requested-With")
			c.Response.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		// if string(c.GetRequest().Method()) == "OPTIONS" {
		// 	c.AbortWithStatus(204) // awalnya 204
		// 	return
		// } // gak bisa
		// if c.Request. /
		if string(c.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(204) 
			// c.JSON(204, "allowed") // harus 204 , kalau diilangin gakbisa
			return
		}
		c.Next(ctx)
	}

	
}
