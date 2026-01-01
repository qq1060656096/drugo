package kernel

import (
	"errors"
)

var (
	ErrServiceNotFound    = errors.New("kernel: service not found")
	ErrKernelNotInContext = errors.New("kernel: kernel not found in context")
	ErrServiceInitFailed  = errors.New("kernel: service initialization failed")
	ErrServiceRunFailed   = errors.New("kernel: service run failed")
	ErrServiceCloseFailed = errors.New("kernel: service close failed")
	ErrServiceType        = errors.New("kernel: service type mismatch")
)

// IsKernelError 判断是否为内核级别的错误（任意一个）
func IsKernelError(err error) bool {
	if err == nil {
		return false
	}
	// 包含所有预定义的内核错误
	kernelErrors := []error{
		ErrServiceNotFound, ErrKernelNotInContext,
		ErrServiceInitFailed, ErrServiceRunFailed, ErrServiceCloseFailed,
		ErrServiceType,
	}
	for _, target := range kernelErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// IsServiceNotFound 判断是否是“服务未找到”错误
func IsServiceNotFound(err error) bool {
	return errors.Is(err, ErrServiceNotFound)
}

func IsServiceInitFailed(err error) bool {
	return errors.Is(err, ErrServiceInitFailed)
}

func IsServiceRunFailed(err error) bool {
	return errors.Is(err, ErrServiceRunFailed)
}

func IsServiceCloseFailed(err error) bool {
	return errors.Is(err, ErrServiceCloseFailed)
}

func IsServiceType(err error) bool {
	return errors.Is(err, ErrServiceType)
}

// Error 是 Drugo 内核的标准错误结构
// 模仿标准库 net.OpError，记录操作名称和原始错误
type Error struct {
	op  string // 发生错误的操作: "service.init", "run", "close", "container.get"
	msg string
	err error // 原始错误
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.err == nil {
		return "kernel: <nil>"
	}
	return "kernel " + e.op + ": " + e.err.Error()
}

// Unwrap 实现 Go 1.13+ 的错误链解包接口
func (e *Error) Unwrap() error {
	return e.err
}

// NewError 创建一个新的内核错误包装
func NewError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		op:  op,
		err: err,
	}
}

func NewServiceNotFound(serviceName string) error {
	return NewError(serviceName, ErrServiceNotFound)
}

func NewServiceInitFailed(serviceName string) error {
	return NewError(serviceName, ErrServiceInitFailed)
}

func NewServiceRunFailed(serviceName string) error {
	return NewError(serviceName, ErrServiceRunFailed)
}

func NewServiceCloseFailed(serviceName string) error {
	return NewError(serviceName, ErrServiceCloseFailed)
}

func NewServiceType(serviceName string) error {
	return NewError(serviceName, ErrServiceType)
}

func NewKernelNotInContext() error {
	return NewError("kernel", ErrKernelNotInContext)
}
