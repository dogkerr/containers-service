// Code generated by hertz generator.

package main

import (
	"dogker/lintang/container-service/biz/dal"
	"dogker/lintang/container-service/biz/router"
	"dogker/lintang/container-service/config"
	"dogker/lintang/container-service/di"
	pb "dogker/lintang/container-service/kitex_gen/container-service/pb/containerservice"
	"net"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/transmeta"
	kitexServer "github.com/cloudwego/kitex/server"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.NewConfig()
	logsCores := initZapLogger(cfg)
	defer logsCores.Sync()
	hlog.SetLogger(logsCores)
	hlog.Info("Halo dunia")

	if err != nil {
		hlog.Fatalf("Config error: %s", err)
	}
	pg := dal.InitPg(cfg) // init postgres & rabbitmq
	rmq := dal.InitRmq(cfg)

	// validation error custom
	customValidationErr := CreateCustomValidationError()

	h := server.Default(server.WithValidateConfig(customValidationErr))
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:6000")
	var opts []kitexServer.Option
	opts = append(opts, kitexServer.WithMetaHandler(transmeta.ServerHTTP2Handler))
	opts = append(opts, kitexServer.WithServiceAddr(addr))

	// ctrKitex := NewContainerService()
	svr := pb.NewServer(new(ContainerServiceImpl), opts...) // kitex rpc server

	go func() {
		err := svr.Run()
		if err != nil {
			hlog.Fatal(err)
		}
	}()

	cSvc := di.InitContainerService(pg, rmq, cfg)

	router.MyRouter(h, cSvc)
	h.Spin()
}

func initZapLogger(cfg *config.Config) *hertzzap.Logger {
	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	productionCfg.EncodeDuration = zapcore.SecondsDurationEncoder
	productionCfg.EncodeCaller = zapcore.ShortCallerEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// log encooder (json for prod, console for dev)
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)
	// loglevel
	logDevLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	logLevelProd := zap.NewAtomicLevelAt(zap.InfoLevel)

	//write sycer
	writeSyncerStdout, writeSyncerFile := getLogWriter(cfg.MaxBackups, cfg.MaxAge)

	prodCfg := hertzzap.CoreConfig{
		Enc: fileEncoder,
		Ws:  writeSyncerFile,
		Lvl: logLevelProd,
	}

	devCfg := hertzzap.CoreConfig{
		Enc: consoleEncoder,
		Ws:  writeSyncerStdout,
		Lvl: logDevLevel,
	}
	logsCores := []hertzzap.CoreConfig{
		prodCfg,
		devCfg,
	}

	prodAndDevLogger := hertzzap.NewLogger(hertzzap.WithZapOptions(zap.WithFatalHook(zapcore.WriteThenPanic)),
		hertzzap.WithCores(logsCores...))
	return prodAndDevLogger
}

func getLogWriter(maxBackup, maxAge int) (writeSyncerStdout zapcore.WriteSyncer, writeSyncerFile zapcore.WriteSyncer) {
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename: "./logs/app.log",

		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	})
	stdout := zapcore.AddSync(os.Stdout)

	return stdout, file
}

type ValidateError struct {
	ErrType string `json:"error_type"`
	 FailField string `json:"validateion_fail_field"`
	 Msg string 	`json:"cause"`
}

// Error implements error interface.
func (e *ValidateError) Error() string {
	if e.Msg != "" {
		return e.ErrType + ": expr_path=" + e.FailField + ", cause=" + e.Msg
	}
	return e.ErrType + ": expr_path=" + e.FailField + ", cause=invalid"
}

func CreateCustomValidationError() *binding.ValidateConfig {
	validateConfig := &binding.ValidateConfig{}
	validateConfig.SetValidatorErrorFactory(func(failField, msg string) error {
		err := ValidateError{
			ErrType:   "validateErr",
			FailField: "[validateFailField]: " + failField,
			Msg:       "[validateErrMsg]: " + msg,
		}

		return &err
	})
	return validateConfig
}
