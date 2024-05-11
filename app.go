package frame

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/ojetra/frame/closer"
	"github.com/ojetra/frame/log"
)

type App struct {
	config Config

	httpRouter *chi.Mux
	httpServer *http.Server

	grpcServer        *grpc.Server
	swaggerHttpServer *http.Server

	closer *closer.Closer

	logger *slog.Logger
}

func New(config Config, logger *slog.Logger) *App {
	return &App{
		config: config,
		closer: closer.New(logger),
		logger: logger,
	}
}

func (x *App) RegisterHttpHandler(method HttpMethod, pattern string, handler HandlerFn) {
	if x.httpRouter == nil {
		x.initHttpRouter()
	}

	innerHandler := func(writer http.ResponseWriter, request *http.Request) {
		response, err := handler(request)
		if err != nil {
			var resultCode int

			if response != nil && response.Code != 0 {
				resultCode = response.Code
			} else {
				resultCode = http.StatusInternalServerError
			}

			setErrorResponse(err, resultCode, writer, request)
			return
		}

		_, err = writer.Write(response.Data)
		if err != nil {
			setErrorResponse(err, http.StatusInternalServerError, writer, request)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}

	switch method {
	case Get:
		x.httpRouter.Get(pattern, innerHandler)
	case Post:
		x.httpRouter.Post(pattern, innerHandler)
	case Head:
		x.httpRouter.Head(pattern, innerHandler)
	case Put:
		x.httpRouter.Put(pattern, innerHandler)
	case Delete:
		x.httpRouter.Delete(pattern, innerHandler)
	case Options:
		x.httpRouter.Options(pattern, innerHandler)
	default:
	}
}

func (x *App) RegisterGrpcHandlers(registrars ...GrpcRegistrar) {
	if x.grpcServer == nil {
		x.initGrpcServer()
	}

	grpcMux := runtime.NewServeMux()

	for _, registrar := range registrars {
		err := registrar.RegisterGrpcHandler(x.grpcServer, grpcMux)
		if err != nil {
			panic(err.Error())
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	x.swaggerHttpServer = &http.Server{
		Handler: mux,
		Addr:    ":8081",
	}
}

func (x *App) Run() {
	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	x.runHttpServer()
	x.runGrpcServer()

	<-notifyCtx.Done()

	x.closer.CloseAll()
}

func (x *App) runHttpServer() {
	if x.httpRouter == nil {
		return
	}

	x.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: x.httpRouter,
	}

	go func() {
		serverErr := x.httpServer.ListenAndServe()
		if serverErr != nil && errors.Is(serverErr, http.ErrServerClosed) {
			return
		}
		log.FatalIfError(serverErr, "cat`t start http server")
	}()

	x.closer.Add(
		func() error {
			x.logger.Info("http: stopping server")

			ctx, cancel := context.WithTimeout(context.Background(), x.config.GetGracefulShutdownTimeoutSecond())
			defer cancel()

			if err := x.httpServer.Shutdown(ctx); err != nil {
				return errors.Wrap(err, "http: failed to stop server")
			}

			return nil
		},
	)
}

func (x *App) runGrpcServer() {
	if x.grpcServer == nil {
		return
	}

	lis, err := net.Listen("tcp", ":8082")
	log.FatalIfError(err, "failed to listen")

	go func() {
		err = x.grpcServer.Serve(lis)
		log.FatalIfError(err, "failed to serve grpc")
	}()

	go func() {
		err = x.swaggerHttpServer.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			return
		}
		log.FatalIfError(err, "cat`t start swagger")
	}()

	x.closer.Add(
		func() error {
			x.logger.Info("grpc: stopping server")

			ctx, cancel := context.WithTimeout(context.Background(), x.config.GetGracefulShutdownTimeoutSecond())
			defer cancel()

			done := make(chan struct{})
			go func() {
				x.grpcServer.GracefulStop()
				close(done)
			}()

			select {
			case <-done:
				x.logger.Warn("grpc: gracefully stopped")
			case <-ctx.Done():
				x.grpcServer.Stop()
				err = fmt.Errorf("error during shutdown server: %w", ctx.Err())

				return errors.Wrap(err, "grpc: force stopped")
			}

			return nil
		},
	)

	x.closer.Add(
		func() error {
			x.logger.Info("swagger: stopping server")

			ctx, cancel := context.WithTimeout(context.Background(), x.config.GetGracefulShutdownTimeoutSecond())
			defer cancel()

			if err = x.swaggerHttpServer.Shutdown(ctx); err != nil {
				return errors.Wrap(err, "swagger: failed to stop server")
			}

			return nil
		},
	)
}

func (x *App) initHttpRouter() {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	x.httpRouter = router
}

func (x *App) initGrpcServer() {
	x.grpcServer = grpc.NewServer()
}
