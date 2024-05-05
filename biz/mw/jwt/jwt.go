package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/hertz-contrib/jwt"
)

var (
	// JwtMiddleware       *jwt.HertzJWTMiddleware
	IdentityKey         = "sub"
	PublicKeyAuthServer = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEnlwXdOFOQFhhEoYksncm/mmRMjVv\nVKiJhzabtB5d2uMV7Xn0SKVzJB4jKUM/05Qcfmxkjt4OyBJNQ4LE5oa3eQ==\n-----END PUBLIC KEY-----\n"
)

type login struct {
	Username string `form:"username,required" json:"username,required"`
	Password string `form:"password,required" json:"password,required"`
}

type User struct {
	ID string
}

func GetJwtMiddleware() *jwt.HertzJWTMiddleware {
	var err error
	publicKeyBlock, _ := pem.Decode([]byte(PublicKeyAuthServer))
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	ECDSAPubKey := publicKey.(*ecdsa.PublicKey)
	JwtMiddleware, err := jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       "dogker digital signature public key auth",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					IdentityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return &User{
				ID: claims[IdentityKey].(string),
			}
		},
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			if v, ok := data.(*User); ok {
				c.Set("userID", v.ID)
				hlog.CtxInfof(ctx, "userId yg request" + v.ID + " " + c.GetString("userID"))

				return true
			}

			return false
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, map[string]interface{}{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer". If you want empty value, use WithoutDefaultTokenHeadName.
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
		KeyFunc: func(t *gojwt.Token) (interface{}, error) {
			// if gojwt.GetSigningMethod(gojwt.SigningMethodECDSA) != t.Method {
			// 	return nil, jwt.ErrInvalidSigningAlgorithm
			// }
			if _, ok := t.Method.(*gojwt.SigningMethodECDSA); !ok {
				return nil, jwt.ErrInvalidSigningAlgorithm
			}

			return ECDSAPubKey, nil
		},
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return JwtMiddleware
}
