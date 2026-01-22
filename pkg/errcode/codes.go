package errcode

import "net/http"

// 错误码规范：
// 第一位表示错误级别：1-系统级，2-通用业务，3-具体业务
// 第二三位表示模块：00-通用，01-用户，02-订单...
// 第四五位表示具体错误

// ============== 成功 ==============

var OK = &Error{Code: 0, Msg: "success", HTTPCode: http.StatusOK}

// ============== 系统级错误 (1xxxx) ==============

var (
	ErrInternal           = NewWithHTTP(10001, "服务内部错误", http.StatusInternalServerError)
	ErrServiceUnavailable = NewWithHTTP(10002, "服务暂不可用", http.StatusServiceUnavailable)
	ErrTimeout            = NewWithHTTP(10003, "请求超时", http.StatusGatewayTimeout)
	ErrTooManyRequests    = NewWithHTTP(10004, "请求过于频繁", http.StatusTooManyRequests)
)

// ============== 通用业务错误 (2xxxx) ==============

var (
	ErrInvalidParam     = NewWithHTTP(20001, "参数错误", http.StatusBadRequest)
	ErrNotFound         = NewWithHTTP(20002, "资源不存在", http.StatusNotFound)
	ErrAlreadyExists    = NewWithHTTP(20003, "资源已存在", http.StatusConflict)
	ErrUnauthorized     = NewWithHTTP(20004, "未授权", http.StatusUnauthorized)
	ErrForbidden        = NewWithHTTP(20005, "禁止访问", http.StatusForbidden)
	ErrValidationFailed = NewWithHTTP(20006, "数据验证失败", http.StatusBadRequest)
	ErrParseFailed      = NewWithHTTP(20007, "数据解析失败", http.StatusBadRequest)
)

// ============== 认证授权 (21xxx) ==============

var (
	ErrTokenInvalid     = NewWithHTTP(21001, "Token 无效", http.StatusUnauthorized)
	ErrTokenExpired     = NewWithHTTP(21002, "Token 已过期", http.StatusUnauthorized)
	ErrTokenMissing     = NewWithHTTP(21003, "Token 缺失", http.StatusUnauthorized)
	ErrPermissionDenied = NewWithHTTP(21004, "权限不足", http.StatusForbidden)
)

// ============== 数据库 (22xxx) ==============

var (
	ErrDatabaseOperation  = NewWithHTTP(22001, "数据库操作失败", http.StatusInternalServerError)
	ErrDatabaseConnection = NewWithHTTP(22002, "数据库连接失败", http.StatusInternalServerError)
	ErrDatabaseDuplicate  = NewWithHTTP(22003, "数据重复", http.StatusConflict)
	ErrDatabaseNotFound   = NewWithHTTP(22004, "数据不存在", http.StatusNotFound)
)

// ============== 用户模块 (301xx) ==============

var (
	ErrUserNotFound      = New(30101, "用户不存在")
	ErrUserAlreadyExists = New(30102, "用户已存在")
	ErrUserPasswordWrong = New(30103, "密码错误")
	ErrUserDisabled      = New(30104, "用户已禁用")
)
