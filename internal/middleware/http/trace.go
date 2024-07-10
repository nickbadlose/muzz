package http

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// NewTraceMiddleware returns a middleware that authenticates the request JWT and sets the user ID on context.
func NewTraceMiddleware(tp trace.TracerProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//if tp == nil {
			//	tp = otel.GetTracerProvider()
			//}
			//ctx, span := tp.Tracer("muzz", trace.WithInstrumentationVersion("0.0")).Start(r.Context(), r.URL.Path)
			//defer span.End()
			//
			//r = r.Clone(ctx) // According to RoundTripper spec, we shouldn't modify the origin request.
			//span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest("muzz-api", r.URL.Path, r)...)

			next.ServeHTTP(w, r)
		})
	}
}
