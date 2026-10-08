package binding

import (
	"errors"
	"net/http"

	"github.com/Dopesonic/catalog-service/internal/pkg/http/httph"
)

type jsonBinding struct{}

func (j jsonBinding) Name() string {
	return "JSON"
}

func (j jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request body")
	}

	if err := httph.DecodeJSON(req, obj); err != nil {
		return err
	}

	return validate(obj)
}
