package middlewares

import (
	"context"
	"modulyn/pkg/db"
	"net/http"

	"github.com/google/uuid"
)

const correlationHeader = "X-Correlation-ID"

// correlationMiddleware creates the middleware handler
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read correlation ID from header
		corrID := r.Header.Get(correlationHeader)

		// If not present, generate new UUID v4
		if corrID == "" {
			corrID = uuid.New().String()
		}

		// Set in response header for client return
		w.Header().Set(correlationHeader, corrID)

		// Store in context for downstream access
		ctx := context.WithValue(r.Context(), db.CorrelationKey, corrID)
		r = r.WithContext(ctx)

		// Proceed to next handler
		next.ServeHTTP(w, r)
	})
}
