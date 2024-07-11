package tracer

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	libVersion = "1.0"
	libName    = "github.com/nickbadlose/muzz/internal/middleware/http"

	// separator for a http dump between the body and metadata.
	separator = "\r\n\r\n"
)

// Attribute returns an attribute.KeyValue from the given interface value.
//
// all cases provided by the otel attribute library are handled. if a type that isn't handled by the lib
// is provided, an error is logged.
func Attribute(key string, value any) (kv attribute.KeyValue) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		return Attribute(key, rv.Elem().Interface())
	}

	switch v := value.(type) {
	case string:
		kv = attribute.String(key, v)
	case []string:
		kv = attribute.StringSlice(key, v)
	case int64:
		kv = attribute.Int64(key, v)
	case int:
		kv = attribute.Int(key, v)
	case float64:
		kv = attribute.Float64(key, v)
	case bool:
		kv = attribute.Bool(key, v)
	case []bool:
		kv = attribute.BoolSlice(key, v)
	case []float64:
		kv = attribute.Float64Slice(key, v)
	case []int64:
		kv = attribute.Int64Slice(key, v)
	case []int:
		kv = attribute.IntSlice(key, v)
	case fmt.Stringer:
		kv = attribute.Stringer(key, v)
	case time.Duration:
		kv = attribute.Float64(key, v.Seconds())
	default:
		log.Printf("unhandled type when formatting interface attribute: %T \n", v)
		return
	}

	return
}

// attributesFromDump splits the metadata and the request body into two separate strings and returns them as
// attributes.
//
// If the `Transfer-Encoding: chunked` header is set and the content length is unknown, then the body will be wrapped
// in content length information. This means the body cannot be prettified by the UI printing it. So it is advised to
// set the Content-Length headers if possible.
func attributesFromRequestDump(data []byte) (attr []attribute.KeyValue) {
	meta, body, split := strings.Cut(string(data), separator)
	if !split {
		attr = append(attr, attribute.String("http.request_dump", meta))
		return
	}

	if len(meta) != 0 {
		attr = append(attr, attribute.String("http.request_metadata", meta))
	}
	if len(body) != 0 {
		attr = append(attr, attribute.String("http.request_body", body))
	}

	return
}
