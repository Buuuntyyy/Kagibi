package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitor struct to hold the rate limiter and the last seen time
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// EndpointLimiter manages rate limits per endpoint
type EndpointLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// sync.Map for better concurrency on visitors
var visitors sync.Map // map[string]*visitor
var endpointLimiter *EndpointLimiter

// Run a background goroutine to remove old entries from the visitors map
func init() {
	go cleanupVisitors()
	endpointLimiter = NewEndpointLimiter()
}

// NewEndpointLimiter creates a limiter with specific limits per endpoint
func NewEndpointLimiter() *EndpointLimiter {
	return &EndpointLimiter{
		limiters: map[string]*rate.Limiter{
			"/auth/login":            rate.NewLimiter(0.1, 5),  // 5 max, 1 every 10s
			"/auth/register":         rate.NewLimiter(0.05, 3), // 3 max, 1 every 20s
			"/users/change-password": rate.NewLimiter(0.02, 2), // 2 max, 1 every 50s
			"/users/profile":         rate.NewLimiter(0.1, 5),  // 5 max, 1 every 10s
			"/files/upload":          rate.NewLimiter(5, 50),   // 50 max, 5 per second (chunked uploads)
			"/files/download":        rate.NewLimiter(10, 100), // 100 max, 10 per second (blob streaming)
			"default":                rate.NewLimiter(10, 30),  // Other endpoints
		},
	}
}

// GetLimiter returns the appropriate limiter for an endpoint (read-optimized)
func (el *EndpointLimiter) GetLimiter(endpoint string) *rate.Limiter {
	el.mu.RLock()
	defer el.mu.RUnlock()

	if limiter, ok := el.limiters[endpoint]; ok {
		return limiter
	}
	return el.limiters["default"]
}

func getVisitor(ip string, endpoint string) *rate.Limiter {
	key := ip + "_" + endpoint

	// Fast path: check if visitor exists
	if val, ok := visitors.Load(key); ok {
		v := val.(*visitor)
		v.lastSeen = time.Now()
		return v.limiter
	}

	// Slow path: create new visitor
	limiter := endpointLimiter.GetLimiter(endpoint)
	newVisitor := &visitor{limiter: limiter, lastSeen: time.Now()}

	// Store or load existing (handles race)
	actual, _ := visitors.LoadOrStore(key, newVisitor)
	return actual.(*visitor).limiter
}

// cleanupVisitors removes visitors that haven't been seen for a while
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		visitors.Range(func(key, value interface{}) bool {
			v := value.(*visitor)
			if time.Since(v.lastSeen) > 3*time.Minute {
				visitors.Delete(key)
			}
			return true
		})
	}
}

// RateLimitMiddleware is the Gin middleware function with endpoint-specific limits
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		endpoint := c.Request.URL.Path
		limiter := getVisitor(ip, endpoint)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}
		c.Next()
	}
}
