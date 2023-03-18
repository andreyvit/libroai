/*
Package httperrors defines an error wrapper that carries extra data for
convenient rendering of errors in HTTP responses, esp. in JSON APIs.
*/
package httperrors

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Error interface {
	error
	HTTPCode() int
	ErrorID() string
	PublicError() string
	ForeachExtra(f func(k string, v interface{}))
}

type BaseError struct {
	id       string
	httpcode int
}

func New(id string, httpcode int) BaseError {
	return BaseError{id, httpcode}
}

func (base BaseError) Error() string {
	return base.String()
}

func (base BaseError) HTTPCode() int {
	return base.httpcode
}

func (base BaseError) ErrorID() string {
	return base.id
}
func (base BaseError) PublicError() string {
	return ""
}
func (base BaseError) ForeachExtra(f func(k string, v interface{})) {
}

func (base BaseError) Is(err error) bool {
	if err == nil {
		return false
	} else if e, ok := err.(Error); ok {
		return e.ErrorID() == base.id && e.HTTPCode() == base.httpcode
	} else {
		return false
	}
}

func (base BaseError) String() string {
	return fmt.Sprintf("%s [HTTP %d]", base.id, base.httpcode)
}

var (
	/*
		Unavailable error should be returned when the service is experiencing
		a temporary downtime (HTTP 500 Internal Server Error).

		An appropriate client behavior is to retry after a delay.

		Examples: critical external API is unreachable, server is out of
		disk space, general unexpected failure.
	*/
	Unavailable = New("unavail", http.StatusInternalServerError)

	/*
		Overload signals expected temporary unavailability of the endpoint
		due to excessive requests or quota violation (HTTP 503 Service Unavailable).

		An appropriate client behavior is to retry with increasing backoff.
	*/
	Overload = New("unavail", http.StatusServiceUnavailable)

	/*
		BadRequest signals a wildly incorrect HTTP call that uses invalid
		parameter names, omits required parameters, uses invalid format for
		parameter values (e.g. a string passed for an integer parameter)
		or uses an unsupported value for a enumerated parameter.

		This should NOT be used for merely invalid data that likely comes from
		the user. Please define an appropriate error value for that.

		An appropriate client behavior is to log an error and fail the operation
		with a generic error message asking to contact support.
	*/
	BadRequest = New("bad_request", http.StatusBadRequest)

	/*
		NotFound signals a that the resource that the HTTP call primarily refers
		to does not exist. E.g. you're trying to update a page that does not exist.

		This should NOT be used for non-existent auxiliary data. For example,
		trying to save an article with a non-existent author should return
		a validation error (which you should define), not NotFound. Similarly,
		an unknown user account on a normal API endpoint should return
		an appropriate authentication error, not NotFound.

		An appropriate client behavior is to remove the UI corresponding to
		the data that this API call operates on.
	*/
	NotFound = New("not_found", http.StatusNotFound)

	/*
		MethodNotAllowed signals a wildly incorrect HTTP call, caused by
		using a mismatched HTTP verb.

		An appropriate client behavior is to log an error and fail the operation
		with a generic error message asking to contact support.
	*/
	MethodNotAllowed = New("bad_request", http.StatusMethodNotAllowed)
)

type DetailedError struct {
	BaseError
	pubmsg string
	cause  error
	extras map[string]interface{}
}

func (err *DetailedError) Unwrap() error {
	return err.cause
}

func (err *DetailedError) Error() string {
	return err.String()
}

func (err *DetailedError) HTTPCode() int {
	return err.httpcode
}

func (err *DetailedError) ErrorID() string {
	return err.id
}

func (err *DetailedError) PublicError() string {
	return err.pubmsg
}

func (err *DetailedError) ForeachExtra(f func(k string, v interface{})) {
	if err.extras != nil {
		for k, v := range err.extras {
			f(k, v)
		}
	}
}

func (base BaseError) Msg(pubmsg string) *DetailedError {
	err := &DetailedError{BaseError: base}
	err.pubmsg = pubmsg
	return err
}
func (err *DetailedError) Msg(pubmsg string) *DetailedError {
	err.pubmsg = pubmsg
	return err
}

func (base BaseError) Wrap(cause error) Error {
	return base.WrapMsg(cause, "")
}

func (base BaseError) WrapMsg(cause error, pubmsg string) Error {
	if cause == nil {
		return nil
	}
	var e Error
	if errors.As(cause, &e) {
		return e // don't re-wrap if already wrapped
	}
	err := &DetailedError{BaseError: base}
	err.pubmsg = pubmsg
	err.cause = cause
	if e, ok := cause.(interface{ HTTPCode() int }); ok {
		err.httpcode = e.HTTPCode()
	} else if e, ok := cause.(interface{ HTTPStatusCode() int }); ok {
		err.httpcode = e.HTTPStatusCode()
	}
	if e, ok := cause.(interface{ ErrorID() string }); ok {
		err.id = e.ErrorID()
	}
	if e, ok := cause.(interface{ PublicError() string }); ok {
		err.pubmsg = e.PublicError()
	}
	return err
}

func (base BaseError) WrapCustom(prototype, cause error) *DetailedError {
	if prototype == nil && cause == nil {
		return nil
	}
	err := &DetailedError{BaseError: base, cause: cause}

	if e, ok := prototype.(interface{ HTTPCode() int }); ok {
		err.httpcode = e.HTTPCode()
	} else if e, ok := prototype.(interface{ HTTPStatusCode() int }); ok {
		err.httpcode = e.HTTPStatusCode()
	}
	if e, ok := prototype.(interface{ ErrorID() string }); ok {
		err.id = e.ErrorID()
	}
	if e, ok := prototype.(interface{ PublicError() string }); ok {
		err.pubmsg = e.PublicError()
	}
	return err
}

func (base BaseError) Extra(k string, v interface{}) *DetailedError {
	return (&DetailedError{BaseError: base}).Extra(k, v)
}

func (err *DetailedError) Extra(k string, v interface{}) *DetailedError {
	if err.extras == nil {
		err.extras = map[string]interface{}{k: v}
	} else {
		err.extras[k] = v
	}
	return err
}

func (err *DetailedError) String() string {
	var buf strings.Builder
	buf.WriteString(err.BaseError.String())
	if err.pubmsg != "" {
		buf.WriteString(" pubmsg=")
		buf.WriteString(strconv.Quote(err.pubmsg))
	}
	err.ForeachExtra(func(k string, v interface{}) {
		buf.WriteByte(' ')
		buf.WriteString(k)
		buf.WriteByte('=')
		fmt.Fprintf(&buf, "%#v", v)
	})
	if err.cause != nil {
		buf.WriteString(" cause: ")
		fmt.Fprintf(&buf, "%+v", err.cause)
	}
	return buf.String()
}

func HTTPMessage(err Error) string {
	if m := err.PublicError(); m != "" {
		return m
	} else {
		return http.StatusText(err.HTTPCode())
	}
}

func ErrorID(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(Error); ok {
		return e.ErrorID()
	} else {
		return ""
	}
}

func HTTPCode(err error) int {
	if err == nil {
		return 0
	} else if e, ok := err.(interface{ HTTPCode() int }); ok {
		return e.HTTPCode()
	} else {
		return http.StatusInternalServerError
	}
}

func Is4xx(err error) bool {
	code := HTTPCode(err)
	return code >= 400 && code <= 499
}

func Is5xx(err error) bool {
	code := HTTPCode(err)
	return code >= 500 && code <= 599
}

const (
	HTTPCodeKey    = "HTTPCode"
	PublicErrorKey = "PublicError"
	ErrorIDKey     = "ErrorID"
)

func value(err Error, key string) interface{} {
	switch key {
	case HTTPCodeKey:
		return err.HTTPCode()
	case PublicErrorKey:
		return err.PublicError()
	case ErrorIDKey:
		return err.ErrorID()
	default:
		var r interface{}
		err.ForeachExtra(func(k string, v interface{}) {
			if k == key {
				r = v
			}
		})
		return r
	}
}

func Value(err error, key string) interface{} {
	if err == nil {
		return ""
	} else {
		var e Error
		if errors.As(err, &e) {
			return value(e, key)
		} else {
			return nil
		}
	}
}

func String(err error, keys ...string) string {
	if err == nil {
		return ""
	} else if e, ok := err.(Error); ok {
		var buf strings.Builder
		for _, k := range keys {
			v := value(e, k)
			if v != nil {
				if buf.Len() > 0 {
					buf.WriteByte(' ')
				}
				buf.WriteString(k)
				buf.WriteRune('=')
				fmt.Fprint(&buf, v)
			}
		}
		return buf.String()
	} else {
		return err.Error()
	}
}
