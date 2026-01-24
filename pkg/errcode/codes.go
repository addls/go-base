package errcode

import "net/http"

// Error code specification:
// The first digit indicates the error level: 1-system, 2-common business, 3-domain-specific business
// The second and third digits indicate the module: 00-common, 01-user, 02-order...
// The fourth and fifth digits indicate the specific error within the module

// ============== Success ==============

var OK = &Error{Code: 0, Msg: "success", HTTPCode: http.StatusOK}

// ============== System errors (1xxxx) ==============

var (
	ErrInternal           = NewWithHTTP(10001, "internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable = NewWithHTTP(10002, "service temporarily unavailable", http.StatusServiceUnavailable)
	ErrTimeout            = NewWithHTTP(10003, "request timeout", http.StatusGatewayTimeout)
	ErrTooManyRequests    = NewWithHTTP(10004, "too many requests", http.StatusTooManyRequests)
)

// ============== Common business errors (2xxxx) ==============

var (
	ErrInvalidParam     = NewWithHTTP(20001, "invalid parameter", http.StatusBadRequest)
	ErrNotFound         = NewWithHTTP(20002, "resource not found", http.StatusNotFound)
	ErrAlreadyExists    = NewWithHTTP(20003, "resource already exists", http.StatusConflict)
	ErrUnauthorized     = NewWithHTTP(20004, "unauthorized", http.StatusUnauthorized)
	ErrForbidden        = NewWithHTTP(20005, "forbidden", http.StatusForbidden)
	ErrValidationFailed = NewWithHTTP(20006, "validation failed", http.StatusBadRequest)
	ErrParseFailed      = NewWithHTTP(20007, "parse failed", http.StatusBadRequest)
)

// ============== Authentication & authorization (21xxx) ==============

var (
	ErrTokenInvalid     = NewWithHTTP(21001, "token is invalid", http.StatusUnauthorized)
	ErrTokenExpired     = NewWithHTTP(21002, "token has expired", http.StatusUnauthorized)
	ErrTokenMissing     = NewWithHTTP(21003, "token is missing", http.StatusUnauthorized)
	ErrPermissionDenied = NewWithHTTP(21004, "permission denied", http.StatusForbidden)
)

// ============== Database (22xxx) ==============

var (
	ErrDatabaseOperation  = NewWithHTTP(22001, "database operation failed", http.StatusInternalServerError)
	ErrDatabaseConnection = NewWithHTTP(22002, "database connection failed", http.StatusInternalServerError)
	ErrDatabaseDuplicate  = NewWithHTTP(22003, "duplicate data", http.StatusConflict)
	ErrDatabaseNotFound   = NewWithHTTP(22004, "data not found", http.StatusNotFound)
)

// ============== User module (301xx) ==============

var (
	ErrUserNotFound      = New(30101, "user not found")
	ErrUserAlreadyExists = New(30102, "user already exists")
	ErrUserPasswordWrong = New(30103, "incorrect password")
	ErrUserDisabled      = New(30104, "user is disabled")
)
