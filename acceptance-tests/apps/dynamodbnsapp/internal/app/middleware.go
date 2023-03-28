package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func tableCtx(next http.Handler) http.Handler {
	return URLSegmentMiddleware(tableNameKey, next)
}

func tableKeyCtx(next http.Handler) http.Handler {
	return URLSegmentMiddleware(tableKeyNameKey, next)
}

func tablePrimaryKeyCtx(next http.Handler) http.Handler {
	return URLSegmentMiddleware(tablePrimaryKeyNameKey, next)
}

func URLSegmentMiddleware(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tableKeyName := chi.URLParam(r, key)
		if tableKeyName == "" {
			_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Invalid request"})
			return
		}
		ctx := context.WithValue(r.Context(), key, tableKeyName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
