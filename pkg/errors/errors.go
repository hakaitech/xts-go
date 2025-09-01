package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different types of errors that can occur
type ErrorType string

const (
	// Authentication errors
	ErrorTypeAuth           ErrorType = "AUTH_ERROR"
	ErrorTypeTokenExpired   ErrorType = "TOKEN_EXPIRED"
	ErrorTypeInvalidCredentials ErrorType = "INVALID_CREDENTIALS"
	
	// API errors
	ErrorTypeAPI            ErrorType = "API_ERROR"
	ErrorTypeRateLimit      ErrorType = "RATE_LIMIT"
	ErrorTypeInvalidRequest ErrorType = "INVALID_REQUEST"
	ErrorTypeServerError    ErrorType = "SERVER_ERROR"
	
	// Network errors
	ErrorTypeNetwork        ErrorType = "NETWORK_ERROR"
	ErrorTypeTimeout        ErrorType = "TIMEOUT"
	ErrorTypeConnection     ErrorType = "CONNECTION_ERROR"
	
	// Configuration errors
	ErrorTypeConfig         ErrorType = "CONFIG_ERROR"
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	
	// Trading specific errors
	ErrorTypeInsufficientFunds ErrorType = "INSUFFICIENT_FUNDS"
	ErrorTypeInvalidOrder      ErrorType = "INVALID_ORDER"
	ErrorTypeOrderRejected     ErrorType = "ORDER_REJECTED"
	ErrorTypeMarketClosed      ErrorType = "MARKET_CLOSED"
	
	// Market Data specific errors
	ErrorTypeInvalidInstrument ErrorType = "INVALID_INSTRUMENT"
	ErrorTypeDataUnavailable   ErrorType = "DATA_UNAVAILABLE"
	ErrorTypeSubscriptionLimit ErrorType = "SUBSCRIPTION_LIMIT"
	
	// WebSocket errors
	ErrorTypeWebSocket      ErrorType = "WEBSOCKET_ERROR"
	ErrorTypeDisconnected   ErrorType = "DISCONNECTED"
)

// XTSError represents a structured error from the XTS API
type XTSError struct {
	Type       ErrorType
	Code       string
	Message    string
	StatusCode int
	Underlying error
	Context    map[string]interface{}
}

// Error implements the error interface
func (e *XTSError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *XTSError) Unwrap() error {
	return e.Underlying
}

// Is implements error checking for errors.Is()
func (e *XTSError) Is(target error) bool {
	if xtse, ok := target.(*XTSError); ok {
		return e.Type == xtse.Type || e.Code == xtse.Code
	}
	return false
}

// AddContext adds context information to the error
func (e *XTSError) AddContext(key string, value interface{}) *XTSError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewXTSError creates a new XTS error
func NewXTSError(errorType ErrorType, code, message string) *XTSError {
	return &XTSError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewAuthError creates a new authentication error
func NewAuthError(message string) *XTSError {
	return NewXTSError(ErrorTypeAuth, "", message)
}

// NewTokenExpiredError creates a new token expired error
func NewTokenExpiredError() *XTSError {
	return NewXTSError(ErrorTypeTokenExpired, "TOKEN_EXPIRED", "Authentication token has expired")
}

// NewAPIError creates a new API error from HTTP response
func NewAPIError(statusCode int, code, message string) *XTSError {
	var errorType ErrorType
	
	switch statusCode {
	case http.StatusUnauthorized:
		errorType = ErrorTypeAuth
	case http.StatusForbidden:
		errorType = ErrorTypeAuth
	case http.StatusBadRequest:
		errorType = ErrorTypeInvalidRequest
	case http.StatusTooManyRequests:
		errorType = ErrorTypeRateLimit
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		errorType = ErrorTypeServerError
	default:
		errorType = ErrorTypeAPI
	}
	
	return &XTSError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Context:    make(map[string]interface{}),
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, underlying error) *XTSError {
	return &XTSError{
		Type:       ErrorTypeNetwork,
		Message:    message,
		Underlying: underlying,
		Context:    make(map[string]interface{}),
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) *XTSError {
	return NewXTSError(ErrorTypeConfig, "", message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *XTSError {
	err := NewXTSError(ErrorTypeValidation, "VALIDATION_ERROR", message)
	err.AddContext("field", field)
	return err
}

// NewTradingError creates a trading-specific error
func NewTradingError(errorType ErrorType, code, message string) *XTSError {
	return NewXTSError(errorType, code, message)
}

// NewWebSocketError creates a WebSocket-specific error
func NewWebSocketError(message string, underlying error) *XTSError {
	return &XTSError{
		Type:       ErrorTypeWebSocket,
		Message:    message,
		Underlying: underlying,
		Context:    make(map[string]interface{}),
	}
}

// IsRetryable determines if an error is retryable
func (e *XTSError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeConnection, ErrorTypeServerError:
		return true
	case ErrorTypeRateLimit:
		return true // Can retry after backoff
	default:
		return false
	}
}

// IsAuthError checks if the error is authentication-related
func (e *XTSError) IsAuthError() bool {
	return e.Type == ErrorTypeAuth || e.Type == ErrorTypeTokenExpired || e.Type == ErrorTypeInvalidCredentials
}

// IsTradingError checks if the error is trading-related
func (e *XTSError) IsTradingError() bool {
	switch e.Type {
	case ErrorTypeInsufficientFunds, ErrorTypeInvalidOrder, ErrorTypeOrderRejected, ErrorTypeMarketClosed:
		return true
	default:
		return false
	}
}

// IsMarketDataError checks if the error is market data-related
func (e *XTSError) IsMarketDataError() bool {
	switch e.Type {
	case ErrorTypeInvalidInstrument, ErrorTypeDataUnavailable, ErrorTypeSubscriptionLimit:
		return true
	default:
		return false
	}
}

// IsWebSocketError checks if the error is WebSocket-related
func (e *XTSError) IsWebSocketError() bool {
	return e.Type == ErrorTypeWebSocket || e.Type == ErrorTypeDisconnected
}

// Helper functions for common errors

// IsTokenExpired checks if the error indicates an expired token
func IsTokenExpired(err error) bool {
	if xtse, ok := err.(*XTSError); ok {
		return xtse.Type == ErrorTypeTokenExpired
	}
	return false
}

// IsRateLimit checks if the error indicates rate limiting
func IsRateLimit(err error) bool {
	if xtse, ok := err.(*XTSError); ok {
		return xtse.Type == ErrorTypeRateLimit
	}
	return false
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if xtse, ok := err.(*XTSError); ok {
		return xtse.IsRetryable()
	}
	return false
}