package rprocessor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/Dopesonic/catalog-service/internal/app/config/section"
	rhandler "github.com/Dopesonic/catalog-service/internal/app/handler/http"
	"github.com/Dopesonic/catalog-service/internal/app/processor"
	"github.com/Dopesonic/catalog-service/internal/app/util"
	"github.com/Dopesonic/catalog-service/internal/pkg/http/httph"
	"github.com/Dopesonic/catalog-service/internal/pkg/http/mzerolog"
)

type httpProc struct {
	server http.Server
	addr   string
}

func NewHTTP(
	hHealth rhandler.Health,
	hCategory rhandler.Category,
	hProduct rhandler.Product,
	cfg section.ProcessorWebServer,
) processor.Processor {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	r.Use(
		httph.NewErrorMiddleware(),
		mzerolog.NewMiddleware(
			mzerolog.WithSkipper(util.IsFilteredHttpRoute),
		),
	)

	vGenericRegHealthCheck(r, hHealth)

	rV1 := r.PathPrefix("/v1").Subrouter()
	v1RegCategoryHandler(rV1, hCategory)
	v1RegProductHandler(rV1, hProduct)

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		log.Info().
			Strs("methods", methods).
			Str("path", path).
			Msg("Route registered")
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Error walking the router")
	}

	p := httpProc{addr: fmt.Sprintf(":%d", cfg.ListenPort)}
	p.server.Handler = r

	return &p
}

func (p *httpProc) serve(l net.Listener) {
	_ = p.server.Serve(l)
}

func (p *httpProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen HTTP")
		return
	}

	log.Info().Str("listen_addr", p.addr).Msg("Listening of TCP addr for HTTP server has been started")

	go p.serve(listener)

	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(listener.Close))

	processor.WatchForShutdown(ctx, wg, processor.NewCloserContextFunc(
		p.server.Shutdown,
		context.Background(),
		5*time.Second,
	))
}
