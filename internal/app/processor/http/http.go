package rprocessor

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Dopesonic/catalog-service/internal/app/config/section"
	rhandler "github.com/Dopesonic/catalog-service/internal/app/handler/http"
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
) *httpProc {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(handlerNotFound)

	vGenericRegHealthCheck(r, hHealth)

	rV1 := r.PathPrefix("/v1").Subrouter()
	v1RegCategoryHandler(rV1, hCategory)
	v1RegProductHandler(rV1, hProduct)

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}
		log.Printf("Route: %s %s ", methods, tpl)
		return nil
	})
	if err != nil {
		log.Printf("Error walking the router: %v", err)
	}

	p := httpProc{addr: fmt.Sprintf(":%d", cfg.ListenPort)}
	p.server.Addr = p.addr
	p.server.Handler = r

	return &p
}

func (p *httpProc) Serve() error {
	log.Printf("Starting HTTP server on %s", p.addr)
	return p.server.ListenAndServe()
}
