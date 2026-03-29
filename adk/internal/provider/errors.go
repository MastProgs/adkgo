package provider

import "fmt"

// AI-NOTE: 모델 API 호출 중 발생할 수 있는 여러 오류 상황을 정의한 Enum입니다.
type ErrorType int

const (
	ErrUnknown ErrorType = iota
	ErrTokenExceeded
	ErrNetwork
	ErrUnauthorized
	ErrRateLimit
	ErrServerError
)

// APIError는 내부 에러 원인과 ErrorType Enum을 함께 갖는 커스텀 에러 구조체입니다.
type APIError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Type, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

// IsRecoverable 함수는 해당 에러가 발생했을 때 예비 키(Fallback Key)로 넘어가서 다시 시도해볼만한 에러인지 판단합니다.
func IsRecoverable(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.Type {
		// AI-NOTE: 토큰 한도 초과, rate limit, 인증 오류일 때만 fallback을 시도합니다.
		case ErrTokenExceeded, ErrRateLimit, ErrUnauthorized:
			return true
		default:
			return false
		}
	}
	// 일반 에러의 경우, 네트워크 에러 등으로 간주하여 재시도하지 않을 수 있으나
	// 정책에 따라 조건부 변경 가능합니다.
	return false
}
