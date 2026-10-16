package binding

import (
	"net/http"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

type queryBinding struct{}

var _ Binding = queryBinding{}

func (queryBinding) Name() string {
	return "URL-QUERY"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	val := req.URL.Query()

	if err := formDecoder.Decode(obj, val); err != nil {
		return &bindingError{msg: err.Error()}
	}

	if err := validate(obj); err != nil {
		return &bindingError{msg: err.Error()}
	}

	return nil
}
