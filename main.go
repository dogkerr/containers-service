// Code generated by hertz generator.

package main

import (
	"dogker/lintang/container-service/biz/dal"
	"dogker/lintang/container-service/biz/router"
	"dogker/lintang/container-service/config"
	"dogker/lintang/container-service/di"
	"dogker/lintang/container-service/kitex_gen/container-service/pb/containergrpcservice"
	"dogker/lintang/container-service/pkg"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/transmeta"
	kitexServer "github.com/cloudwego/kitex/server"

	"github.com/hertz-contrib/cors"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"

	"github.com/hertz-contrib/swagger"     // hertz-swagger middleware
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	// hertz-swagger middleware
	// swagger embed files
)

// @title go-container-service-lintang
// @version 1.0
// @description init container service buat dogker

// @contact.name lintang
// @description container service dogker

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 103.175.219.0:8888
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.NewConfig()
	logsCores := pkg.InitZapLogger(cfg)
	defer logsCores.Sync()
	hlog.SetLogger(logsCores)

	if err != nil {
		hlog.Fatalf("Config error: %s", err)
	}
	pg := dal.InitPg(cfg) // init postgres & rabbitmq
	rmq := dal.InitRmq(cfg)

	cc, err := grpc.NewClient(cfg.GRPC.MonitorURL+"?wait=30s", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.L().Fatal("Newclient gprc (main)", zap.Error(err))
	}
	// validation error custom
	customValidationErr := pkg.CreateCustomValidationError()

	h := server.Default(server.WithValidateConfig(customValidationErr), server.WithExitWaitTime(4*time.Second), server.WithRedirectTrailingSlash(false))
	// h.Use(Cors())

	h.Use(pkg.AccessLog())

	corsHandler := cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "content-type", "authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"},
		ExposeHeaders:    []string{"Origin", "content-type", "authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"},
		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}) // ini gakbisa cok

	h.Use(corsHandler)

	var callback []route.CtxCallback
	callback = append(callback, rmq.Close, pg.ClosePostgres)
	h.Engine.OnShutdown = append(h.Engine.OnShutdown, callback...) /// graceful shutdown
	cSvc := di.InitContainerService(pg, rmq, cfg, cc)

	// swagger
	url := swagger.URL("http://localhost:8888/swagger/doc.json") // The url pointing to API definition
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))

	// kitex grpc server
	var opts []kitexServer.Option
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(`0.0.0.0:%s`, cfg.GRPC.URLGrpc)) // grpc address

	opts = append(opts, kitexServer.WithMetaHandler(transmeta.ServerHTTP2Handler))
	opts = append(opts, kitexServer.WithServiceAddr(addr))
	opts = append(opts, kitexServer.WithExitWaitTime(5*time.Second))
	opts = append(opts, kitexServer.WithGRPCReadBufferSize(1024*1024*100))

	opts = append(opts, kitexServer.WithGRPCWriteBufferSize(1024*1024*100))
	opts = append(opts, kitexServer.WithGRPCInitialConnWindowSize(1024*1024*100))
	opts = append(opts, kitexServer.WithGRPCInitialWindowSize(1024*1024*100))
	opts = append(opts, kitexServer.WithGRPCMaxHeaderListSize(1024*1024*100))

	cGrpcSvc := di.InitContainerGRPCService(pg, rmq, cfg, cc)
	srv := containergrpcservice.NewServer(cGrpcSvc, opts...)
	klog.SetLogger(pkg.InitZapKitexLogger(cfg))

	go func() {
		err := srv.Run()
		if err != nil {
			zap.L().Fatal("srv.Run()", zap.Error(err))
		}
	}()

	router.MyRouter(h, cSvc)
	h.Spin()
}
