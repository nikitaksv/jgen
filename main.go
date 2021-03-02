package main

import (
	"context"
	"encoding/json"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
	"github.com/micro/micro/v3/service/server"
	"github.com/nikitaksv/gendata/handler"
	pb "github.com/nikitaksv/gendata/proto"
	"github.com/nikitaksv/gendata/zapLogger"
	"go.uber.org/zap"
)

func main() {
	log, _ := zapLogger.NewLogger()
	logger.DefaultLogger = log

	// Create service
	srv := service.New(
		service.Name("go.micro.srv.gendata"),
		service.Version("latest"),
		service.WrapHandler(logWrapper(log)),
	)

	// Register handler
	_ = pb.RegisterGendataHandler(srv.Server(), handler.New(log))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

func logWrapper(log logger.Logger) func(server.HandlerFunc) server.HandlerFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			err := next(ctx, req, rsp)
			reqBs, _ := json.Marshal(req.Body())
			rspBs, _ := json.Marshal(rsp)
			log.Log(
				logger.InfoLevel,
				"logWrapper",
				zap.String("request", string(reqBs)),
				zap.String("response", string(rspBs)),
				zap.Error(err),
			)
			return err
		}
	}
}
