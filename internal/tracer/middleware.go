package tracer

import (
	"fmt"
	"net/http"

	"github.com/nickbadlose/muzz/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http/httputil"
)

const (
	tracingHeader = "otel-trace-id"
	httpSpanName  = "HTTP"
)

// HTTPDebugMiddleware returns a middleware that traces all request and response information for the server.
//
// This should not be used in production.
func HTTPDebugMiddleware(tp trace.TracerProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tp == nil {
				tp = otel.GetTracerProvider()
			}

			tr := tp.Tracer(libName, trace.WithInstrumentationVersion(libVersion))
			ctx, span := tr.Start(
				r.Context(),
				fmt.Sprintf("%s %s %s", httpSpanName, r.Method, r.URL.String()),
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			// continue without tracing if it isn't configured.
			if !span.IsRecording() || !span.SpanContext().IsValid() {
				next.ServeHTTP(w, r)
				return
			}

			r = r.Clone(ctx)
			span.SetAttributes(semconv.HTTPClientAttributesFromHTTPRequest(r)...)
			span.SetAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...)
			span.SetAttributes(Attribute("http.query_params", r.URL.RawQuery))

			reqDump, err := httputil.DumpRequest(r, true)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				logger.Error(ctx, "http request dump failed", err, zap.String("url", r.URL.String()))
			}
			if err == nil {
				span.SetAttributes(attributesFromRequestDump(reqDump)...)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPResponseHeaders sets any response headers we wish to apply to a request.
func HTTPResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get(tracingHeader) == "" {
			span := trace.SpanFromContext(r.Context())
			traceID := span.SpanContext().TraceID().String()

			w.Header().Add(
				tracingHeader,
				traceID,
			)
		}

		next.ServeHTTP(w, r)
	})
}
