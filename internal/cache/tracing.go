package cache

import (
	"fmt"
	"github.com/mediocregopher/radix/v4"
	"github.com/nickbadlose/muzz/internal/tracer"
	"go.opentelemetry.io/otel/trace"
)

const (
	libVersion = "1.0"
	libName    = "github.com/nickbadlose/muzz/internal/cache"

	redisCommand = "redis.command"
	redisKey     = "redis.key"
	redisData    = "redis.data"
	redisArgs    = "redis.args"
)

// getKey from the generic list of args passed to Cache.Do.
func getKey(args []any) string {
	if len(args) == 0 {
		return ""
	}

	key, ok := args[0].(string)
	if !ok {
		return ""
	}

	return key
}

// setTraceAttributes sets all argument attributes on the provided span.
//
// this will add all data to the trace, so do not use in production.
func setTraceAttributes(span trace.Span, args []any) {
	if len(args) == 0 {
		return
	}

	// TODO check this with setEx by clearing cache

	// omit the key (first arg) as we always trace this at a higher level.
	for i, arg := range args[1:] {
		span.SetAttributes(tracer.Attribute(
			fmt.Sprintf("%s.%d", redisArgs, i),
			arg,
		))
	}
}

func setDataAttribute(span trace.Span, rcv any) {
	if rcv == nil {
		return
	}

	maybe, ok := rcv.(*radix.Maybe)
	if ok {
		if !maybe.Empty {
			span.SetAttributes(tracer.Attribute(redisData, maybe.Rcv))
		}
		return
	}

	span.SetAttributes(tracer.Attribute(redisData, rcv))
}
