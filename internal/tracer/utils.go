package tracer

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

const (
	libVersion = "1.0"
	libName    = "github.com/nickbadlose/muzz/internal/tracer"

	// separator for a http dump between the body and metadata.
	separator = "\r\n\r\n"
)

// Attribute returns an attribute.KeyValue from the given interface value.
//
// all cases provided by the otel attribute library are handled. if a type that isn't handled by the lib
// is provided, an error is logged.
func Attribute(key string, value any) attribute.KeyValue {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		return Attribute(key, rv.Elem().Interface())
	}

	var kv attribute.KeyValue
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
	case time.Duration:
		kv = attribute.Float64(key, v.Seconds())
	case fmt.Stringer:
		kv = attribute.Stringer(key, v)
	default:
		log.Printf("unhandled type when formatting interface attribute: %T \n", v)
		return kv
	}

	return kv
}

// attributesFromRequestDump splits the metadata and the request body into two separate strings and returns them as
// attributes.
//
// If the `Transfer-Encoding: chunked` header is set and the content length is unknown, then the body will be wrapped
// in content length information. This means the body cannot be prettified by the UI printing it. So it is advised to
// set the Content-Length headers if possible.
func attributesFromRequestDump(data []byte) []attribute.KeyValue {
	attr := make([]attribute.KeyValue, 0, 2)
	meta, body, split := strings.Cut(string(data), separator)
	if !split {
		attr = append(attr, attribute.String("http.request_dump", meta))
		return attr
	}

	if meta != "" {
		attr = append(attr, attribute.String("http.request_metadata", meta))
	}
	if body != "" {
		attr = append(attr, attribute.String("http.request_body", body))
	}

	return attr
}
