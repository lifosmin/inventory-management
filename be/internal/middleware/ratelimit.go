package middleware

import (
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if time.Since(v.lastSeen) > window {
				v.count = 1
				v.lastSeen = time.Now()
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			v.count++
			v.lastSeen = time.Now()
			if v.count > maxRequests {
				mu.Unlock()
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
