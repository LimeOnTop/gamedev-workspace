package apperr

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DevMode controls whether clients receive detailed internal error strings.
// Default false (production-safe). Set via Configure from config DEV_MODE.
var DevMode bool

const (
	MsgInvalidRequest = "invalid request"
	MsgInternal       = "internal server error"
	MsgNotFound       = "not found"
	MsgTooLarge       = "file is too large"
)

var (
	ErrNotFound   = errors.New(MsgNotFound)
	ErrValidation = errors.New("validation failed")
	ErrTooLarge   = errors.New(MsgTooLarge)
)

func Configure(devMode bool) {
	DevMode = devMode
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return ErrValidation }

// Validation returns a business error whose message is safe to show to clients.
func Validation(format string, args ...any) error {
	return &validationError{msg: fmt.Sprintf(format, args...)}
}

// Status maps an error to an HTTP status code.
func Status(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrTooLarge):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// Message returns a client-safe message for err. The real error is always logged
// for internal failures; business errors keep their own text.
func Message(err error) string {
	var verr *validationError
	switch {
	case errors.As(err, &verr):
		return verr.msg
	case errors.Is(err, ErrNotFound):
		return MsgNotFound
	case errors.Is(err, ErrTooLarge):
		return MsgTooLarge
	}
	log.Printf("internal error: %v", err)
	if DevMode {
		return err.Error()
	}
	raw := strings.TrimSpace(err.Error())
	if raw != "" && !looksInternal(raw) {
		return raw
	}
	return MsgInternal
}

// Respond writes err as a JSON error response.
func Respond(c *gin.Context, err error) {
	c.JSON(Status(err), gin.H{"error": Message(err)})
}

// Bind responds to request validation/binding failures.
func Bind(c *gin.Context, err error) {
	if DevMode {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": MsgInvalidRequest})
}

func looksInternal(msg string) bool {
	lower := strings.ToLower(msg)
	markers := []string{
		"pq:",
		"sql:",
		"redis",
		"connection refused",
		"dial tcp",
		"i/o timeout",
		"context deadline",
		"context canceled",
		"driver:",
		"begin tx",
		"commit tx",
		"panic:",
		"goroutine",
		"/users/",
		"/home/",
		"/data/",
		"password=",
		"postgres://",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}
