package mvp

import (
	"net/http"

	"github.com/andreyvit/buddyd/internal/httperrors"
)

var (
	ErrTooManyRequests = httperrors.New("too_many_requests", http.StatusTooManyRequests)
)
