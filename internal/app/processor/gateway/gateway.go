package pgateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Dopesonic/catalog-service/internal/app/config/section"
	"github.com/Dopesonic/catalog-service/internal/app/processor"
	catalogv1 "github.com/Dopesonic/catalog-service/internal/pkg/grpc/gen/catalog/v1"
)

const (
	gatewayReadHeaderTimeout = 5 * time.Second
	gatewayReadTimeout       = 30 * time.Second
	gatewayWriteTimeout      = 30 * time.Second
	gatewayIdleTimeout       = 120 * time.Second
	gatewayShutdownTimeout   = 5 * time.Second
)

type gatewayProc struct {
	server   http.Server
	addr     string
	grpcAddr string
}

func NewGateway(
	cfgGateway section.ProcessorGateway,
	cfgGrpc section.ProcessorGrpc,
) processor.Processor {
	gatewayAddr := fmt.Sprintf(":%d", cfgGateway.ListenPort)
	grpcAddr := net.JoinHostPort("localhost", strconv.Itoa(int(cfgGrpc.ListenPort)))

	return &gatewayProc{
		server: http.Server{
			Addr:              gatewayAddr,
			ReadHeaderTimeout: gatewayReadHeaderTimeout,
			ReadTimeout:       gatewayReadTimeout,
			WriteTimeout:      gatewayWriteTimeout,
			IdleTimeout:       gatewayIdleTimeout,
		},
		addr:     gatewayAddr,
		grpcAddr: grpcAddr,
	}
}

func (p *gatewayProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, p.grpcAddr, opts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register Gateway routes")
		return
	}

	p.server.Handler = mux

	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen Gateway HTTP")
		return
	}

	log.Info().
		Str("listen_addr", p.addr).
		Str("grpc_addr", p.grpcAddr).
		Msg("Gateway listener started")

	go func() {
		_ = p.server.Serve(listener)
	}()

	processor.WatchForShutdown(
		ctx,
		wg,
		processor.CloserFunc(listener.Close),
	)

	processor.WatchForShutdown(
		ctx,
		wg,
		processor.NewCloserContextFunc(
			p.server.Shutdown,
			context.Background(),
			gatewayShutdownTimeout,
		),
	)
}
