package errs

import (
	"errors"

	"github.com/gitslim/go-ragger/internal/web/tpl"
	datastar "github.com/starfederation/datastar/sdk/go"
)

var (
	ErrInternal       = errors.New("Ошибка сервера")
	ErrBadCredentials = errors.New("Неправильные учетные данные")
)

func ShowErrors(sse *datastar.ServerSentEventGenerator, errs ...error) {
	sse.MergeFragmentTempl(tpl.ErrorMessages(errs...))
}
