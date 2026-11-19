package mmetric

import (
	"errors"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
)

var allowedReasons = map[entity.ErrorKind]struct{}{
	entity.KindUnknown:            {},
	entity.KindNotFound:           {},
	entity.KindAlreadyExists:      {},
	entity.KindInvalidArgument:    {},
	entity.KindFailedPrecondition: {},
}

func createReason(err error) string {
	if err == nil {
		return createReasonNone
	}

	var appErr *entity.AppError
	if errors.As(err, &appErr) {
		return string(normalizeReason(appErr.Kind))
	}

	var httpCoder interface{ HTTPStatus() int }
	if errors.As(err, &httpCoder) {
		kind := entity.KindFromHTTPStatus(httpCoder.HTTPStatus())
		return string(normalizeReason(kind))
	}

	return string(entity.KindUnknown)
}

func normalizeReason(kind entity.ErrorKind) entity.ErrorKind {
	if _, ok := allowedReasons[kind]; !ok {
		return entity.KindUnknown
	}

	return kind
}
