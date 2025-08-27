package retry

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/corynth/corynth/pkg/logging"
)

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// Config contains retry configuration options
type Config struct {
	// MaxAttempts is the maximum number of retry attempts (default: 3)
	MaxAttempts int
	
	// InitialDelay is the initial delay between retries (default: 1s)
	InitialDelay time.Duration
	
	// MaxDelay is the maximum delay between retries (default: 30s)
	MaxDelay time.Duration
	
	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
	
	// Jitter adds randomness to delays to prevent thundering herd (default: true)
	Jitter bool
	
	// RetryableErrors is a list of error messages that should trigger retries
	RetryableErrors []string
	
	// Logger for retry operations
	Logger *logging.Logger
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:       3,
		InitialDelay:      time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
		RetryableErrors:   []string{"timeout", "connection", "network", "temporary"},
		Logger:            logging.NewDefaultLogger("retry"),
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func (c Config) WithMaxAttempts(attempts int) Config {
	c.MaxAttempts = attempts
	return c
}

// WithInitialDelay sets the initial delay between retries
func (c Config) WithInitialDelay(delay time.Duration) Config {
	c.InitialDelay = delay
	return c
}

// WithMaxDelay sets the maximum delay between retries
func (c Config) WithMaxDelay(delay time.Duration) Config {
	c.MaxDelay = delay
	return c
}

// WithBackoffMultiplier sets the exponential backoff multiplier
func (c Config) WithBackoffMultiplier(multiplier float64) Config {
	c.BackoffMultiplier = multiplier
	return c
}

// WithJitter enables or disables jitter
func (c Config) WithJitter(enabled bool) Config {
	c.Jitter = enabled
	return c
}

// WithRetryableErrors sets the list of error patterns that should trigger retries
func (c Config) WithRetryableErrors(errors []string) Config {
	c.RetryableErrors = errors
	return c
}

// WithLogger sets a custom logger
func (c Config) WithLogger(logger *logging.Logger) Config {
	c.Logger = logger
	return c
}

// Execute executes a function with retry logic
func Execute(ctx context.Context, config Config, fn RetryableFunc) error {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 1
	}
	
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 1 {
				config.Logger.Info("operation succeeded after %d attempts", attempt)
			}
			return nil
		}
		
		lastErr = err
		
		// Check if this is the last attempt
		if attempt == config.MaxAttempts {
			config.Logger.Error("operation failed after %d attempts: %v", config.MaxAttempts, err)
			break
		}
		
		// Check if error is retryable
		if !isRetryableError(err, config.RetryableErrors) {
			config.Logger.Warn("non-retryable error encountered: %v", err)
			return err
		}
		
		config.Logger.Warn("attempt %d failed, retrying in %v: %v", attempt, delay, err)
		
		// Wait for the delay with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		
		// Calculate next delay with exponential backoff
		delay = calculateNextDelay(delay, config)
	}
	
	return fmt.Errorf("operation failed after %d attempts, last error: %w", config.MaxAttempts, lastErr)
}

// ExecuteWithTimeout executes a function with both retry logic and a timeout
func ExecuteWithTimeout(timeout time.Duration, config Config, fn RetryableFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return Execute(ctx, config, fn)
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error, retryablePatterns []string) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// calculateNextDelay calculates the next delay using exponential backoff with jitter
func calculateNextDelay(currentDelay time.Duration, config Config) time.Duration {
	// Apply exponential backoff
	nextDelay := time.Duration(float64(currentDelay) * config.BackoffMultiplier)
	
	// Apply maximum delay cap
	if nextDelay > config.MaxDelay {
		nextDelay = config.MaxDelay
	}
	
	// Apply jitter to prevent thundering herd
	if config.Jitter {
		// Add ±25% jitter
		jitterRange := float64(nextDelay) * 0.25
		jitter := (math.Sin(float64(time.Now().UnixNano())) + 1) * jitterRange / 2
		nextDelay = time.Duration(float64(nextDelay) + jitter - jitterRange/2)
		
		// Ensure we don't go below minimum or above maximum
		if nextDelay < config.InitialDelay {
			nextDelay = config.InitialDelay
		}
		if nextDelay > config.MaxDelay {
			nextDelay = config.MaxDelay
		}
	}
	
	return nextDelay
}

// contains is a case-insensitive string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}