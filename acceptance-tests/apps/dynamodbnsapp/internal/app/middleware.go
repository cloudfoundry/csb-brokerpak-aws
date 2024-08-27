package app

import (
	"context"
	"fmt"
	"net/http"
)

func tableCtx(next http.HandlerFunc) http.HandlerFunc {
	return URLSegmentMiddleware(tableNameKey, next)
}

func tableKeyCtx(next http.HandlerFunc) http.HandlerFunc {
	return URLSegmentMiddleware(tableKeyNameKey, next)
}

func tablePrimaryKeyCtx(next http.HandlerFunc) http.HandlerFunc {
	return URLSegmentMiddleware(tablePrimaryKeyNameKey, next)
}

func URLSegmentMiddleware(key string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyValue := r.PathValue(key)
		if keyValue == "" {
			writeJSONResponse(w, http.StatusBadRequest, NewErrResponse(fmt.Errorf("invalid request - empty key %s", key)))
			return
		}
		ctx := context.WithValue(r.Context(), key, keyValue)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
