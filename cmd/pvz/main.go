package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	pvz_service "pvz/internal/app/grpc"
	"pvz/internal/auth"
	"pvz/internal/config"
	"pvz/internal/dummyUUIDGenerator"
	"pvz/internal/jaeger"
	"pvz/internal/logger"
	"pvz/internal/mw"
	"pvz/internal/postgres"
	auth_service "pvz/internal/usecase/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
	"pvz/pkg/closer"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	c := closer.NewCloser(syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	c.Add(func() error {
		cancel()
		return nil
	})

	authCore := auth.New(cfg.Auth)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.PanicInterceptor,
			mw.ErrorInterceptor(),
			mw.MetricInterceptor(),
			mw.AuthInterceptor(authCore),
		),
	)
	c.Add(func() error {
		grpcServer.GracefulStop()
		return nil
	})

	conn, err := postgres.Connect(ctx, cfg.Dsn)
	if err != nil {
		log.Fatalf("db connection failed: %s", err)
	}

	pvzService := pvz.New(pvz.Deps{
		Repo:          conn,
		UUIDGenerator: &dummyUUIDGenerator.Generator{},
	})
	authService := auth_service.NewAuthService(auth_service.Deps{
		Issuer: authCore,
		Repo:   conn,
	})

	service := pvz_service.NewService(
		cfg.DateTimeFormat,
		pvz_service.Deps{
			PVZ:  pvzService,
			Auth: authService,
		})

	tracer, tracerCloser, err := jaeger.InitJaeger(&cfg.Jaeger)
	if err != nil {
		log.Fatalf("init jaeger failed: %s", err)
	}
	c.Add(tracerCloser.Close)

	opentracing.SetGlobalTracer(tracer)
	pvzpb.RegisterPVZServiceServer(grpcServer, service)
	reflection.Register(grpcServer)

	port := cfg.GRPCServer.Port
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen tcp: %s", err)
	}

	go runGrpcServer(cfg.GRPCServer, tcpListener, grpcServer)

	gwMux, httpCloser, err := registerGatewayMux(tcpListener)
	if err != nil {
		log.Fatal(err.Error())
	}
	c.Add(httpCloser)

	httpPort := cfg.HTTPServer.Port
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: gwMux,
	}
	go runHttpServer(cfg.HTTPServer, httpServer)

	obsMux := http.NewServeMux()
	obsMux.Handle("/metrics", promhttp.Handler())
	obsServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.OBSServer.Port),
		Handler: obsMux,
	}
	c.Add(obsServer.Close)
	go runHttpServer(cfg.OBSServer, obsServer)

	<-ctx.Done()
	logger.Info("grpc and http servers shut down gracefully")
}

func runGrpcServer(cfg config.GRPCServer, listener net.Listener, grpcServer *grpc.Server) {
	logger.Infof("running grpc server on port %s\n", cfg.Port)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatalf("failed to serve grpc server: %s", err)
	}
}

func runHttpServer(cfg config.HTTPServer, httpServer *http.Server) {
	logger.Debug(fmt.Sprintf("running http server on port %s\n", cfg.Port))

	if err := httpServer.ListenAndServe(); err != nil {
		logger.Error(fmt.Errorf("failed to serve http server: %w", err))
	}
}

func registerGatewayMux(tcpListener net.Listener) (*http.ServeMux, func() error, error) {
	gatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(mw.ForwardResponseFunc),
	)

	conn, err := grpc.NewClient(
		tcpListener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to new grpc client: %w", err)
	}

	if err := pvzpb.RegisterPVZServiceHandler(context.Background(), gatewayMux, conn); err != nil {
		return nil, nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gatewayMux)

	return mux, conn.Close, err
}
