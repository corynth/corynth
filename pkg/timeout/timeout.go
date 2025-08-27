package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/corynth/corynth/pkg/logging"
)

// Config contains timeout configuration
type Config struct {
	// Default timeout for operations (default: 30s)
	DefaultTimeout time.Duration
	
	// Maximum allowed timeout (default: 300s)
	MaxTimeout time.Duration
	
	// Logger for timeout operations
	Logger *logging.Logger
}

// DefaultConfig returns a sensible default timeout configuration
func DefaultConfig() Config {
	return Config{
		DefaultTimeout: 30 * time.Second,
		MaxTimeout:     300 * time.Second,
		Logger:         logging.NewDefaultLogger("timeout"),
	}
}

// WithContext creates a context with timeout from the provided parameters
func WithContext(ctx context.Context, params map[string]interface{}, config Config) (context.Context, context.CancelFunc) {
	timeout := config.DefaultTimeout
	
	// Check if timeout is specified in parameters
	if timeoutParam, ok := params["timeout"]; ok {
		if timeoutNum, ok := timeoutParam.(float64); ok && timeoutNum > 0 {
			requestedTimeout := time.Duration(timeoutNum) * time.Second
			
			// Enforce maximum timeout limit
			if requestedTimeout > config.MaxTimeout {
				config.Logger.Warn("requested timeout %v exceeds maximum %v, using maximum", 
					requestedTimeout, config.MaxTimeout)
				timeout = config.MaxTimeout
			} else {
				timeout = requestedTimeout
			}
		}
	}
	
	config.Logger.Debug("setting operation timeout to %v", timeout)
	return context.WithTimeout(ctx, timeout)
}

// ExecuteWithTimeout executes a function with timeout based on parameters
func ExecuteWithTimeout(ctx context.Context, params map[string]interface{}, config Config, fn func(context.Context) error) error {
	timeoutCtx, cancel := WithContext(ctx, params, config)
	defer cancel()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Execute the function in a goroutine
	go func() {
		resultChan <- fn(timeoutCtx)
	}()
	
	// Wait for either completion or timeout
	select {
	case err := <-resultChan:
		return err
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			timeout := config.DefaultTimeout
			if timeoutParam, ok := params["timeout"]; ok {
				if timeoutNum, ok := timeoutParam.(float64); ok && timeoutNum > 0 {
					timeout = time.Duration(timeoutNum) * time.Second
				}
			}
			config.Logger.Error("operation timed out after %v", timeout)
			return fmt.Errorf("operation timed out after %v", timeout)
		}
		return timeoutCtx.Err()
	}
}

// ValidateTimeout validates timeout parameter in plugin parameters
func ValidateTimeout(params map[string]interface{}, config Config) error {
	if timeout, ok := params["timeout"]; ok {
		if timeoutNum, ok := timeout.(float64); !ok {
			return fmt.Errorf("timeout must be a number")
		} else if timeoutNum <= 0 {
			return fmt.Errorf("timeout must be positive")
		} else if time.Duration(timeoutNum)*time.Second > config.MaxTimeout {
			return fmt.Errorf("timeout cannot exceed %v seconds", config.MaxTimeout.Seconds())
		}
	}
	return nil
}

// ParseTimeout extracts and validates timeout from parameters
func ParseTimeout(params map[string]interface{}, defaultTimeout time.Duration) (time.Duration, error) {
	if timeoutParam, ok := params["timeout"]; ok {
		if timeoutNum, ok := timeoutParam.(float64); ok {
			if timeoutNum <= 0 {
				return 0, fmt.Errorf("timeout must be positive, got %v", timeoutNum)
			}
			if timeoutNum > 3600 { // 1 hour max
				return 0, fmt.Errorf("timeout cannot exceed 3600 seconds, got %v", timeoutNum)
			}
			return time.Duration(timeoutNum) * time.Second, nil
		}
		return 0, fmt.Errorf("timeout must be a number, got %T", timeoutParam)
	}
	
	return defaultTimeout, nil
}

// WithDeadline creates a context with deadline for operations that don't use parameters
func WithDeadline(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}