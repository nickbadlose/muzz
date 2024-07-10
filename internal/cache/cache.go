package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"

	"github.com/mediocregopher/radix/v4"
)

const (
	get   = "GET"
	setEx = "SETEX"

	connectionPoolSize = 10
)

var (
	// ErrCacheMiss error is returned when we do not find anything in the cache with the given key.
	ErrCacheMiss = errors.New("cache entry does not exist with the given key")
	// errNoConnection error returned when there is no connection initialised.
	errNoConnection = errors.New(`no cache connection`)
)

type (
	// Duration represents a time.Duration which can be marshalled into a format redis understands
	// for setting expiry times in seconds.
	Duration time.Duration
)

// MarshalBinary implements BinaryMarshaller interface.
func (d Duration) MarshalBinary() (data []byte, err error) {
	return json.Marshal(time.Duration(d).Seconds())
}

// Cache struct is responsible for interacting with the redis cache.
type Cache struct {
	client radix.Client
	config *config
}

// isDebugEnabled returns whether the cache is configured to enable debugging mode.
func (c *Cache) isDebugEnabled() bool {
	if c.config == nil {
		return false
	}
	return c.config.debugEnabled
}

// getTracer returns a tracer configured with the cache trace name.
func (c *Cache) getTracer() trace.Tracer {
	tp := otel.GetTracerProvider()
	if c.config != nil && tp == nil {
		tp = c.config.tracerProvider
	}
	return tp.Tracer(libName, trace.WithInstrumentationVersion(libVersion))
}

// Credentials for connecting to the cache instance.
type Credentials struct {
	// Host of the cache to connect to.
	Host string
	// Password to authenticate with.
	Password string
}

// New connect to redis and returns a new instance of the Cache to interact with.
func New(ctx context.Context, c *Credentials, opts ...Option) (*Cache, error) {
	if c == nil {
		return nil, errors.New("credentials must be provided")
	}

	cfg := &config{
		connectionPoolSize: connectionPoolSize,
		debugEnabled:       false,
		// returns the global tracer provider or a noop if none is set.
		tracerProvider: otel.GetTracerProvider(),
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	client, err := (radix.PoolConfig{
		Dialer: radix.Dialer{
			AuthPass: c.Password,
		},
		Size: cfg.connectionPoolSize,
	}).New(ctx, "tcp", c.Host)

	if err != nil {
		return nil, err
	}

	return &Cache{client: client, config: cfg}, nil
}

// Client returns the underlying radix redis client.
func (c *Cache) Client() (radix.Client, error) {
	if c.client == nil {
		return nil, errNoConnection
	}
	return c.client, nil
}

// Close closes the underlying connection to the redis server.
func (c *Cache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Get attempts to read data from cache and unmarshalls it into rcv.
func (c *Cache) Get(ctx context.Context, key string, rcv any) error {
	var body string
	maybe := &radix.Maybe{Rcv: &body}

	err := c.Do(ctx, maybe, get, key)
	if err != nil {
		return err
	}

	if maybe.Null {
		return ErrCacheMiss
	}

	err = json.Unmarshal([]byte(body), rcv)

	return err
}

// SetEx writes into the cache with a custom expiration time.
func (c *Cache) SetEx(ctx context.Context, key string, val any, expirationDuration time.Duration) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return c.Do(ctx, nil, setEx, key, Duration(expirationDuration), string(body))
}

// Do builds a generic radix.Action command and runs it.
func (c *Cache) Do(ctx context.Context, rcv any, cmd string, args ...any) error {
	if c.client == nil {
		return errNoConnection
	}

	action := radix.FlatCmd(rcv, cmd, args...)

	tr := c.getTracer()
	ctx, span := tr.Start(
		ctx,
		fmt.Sprintf("%s/%s", redisCommand, cmd),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// continue without tracing if none is set.
	if !span.IsRecording() || !span.SpanContext().IsValid() {
		return c.client.Do(ctx, action)
	}

	span.SetAttributes(attribute.String(redisCommand, cmd), attribute.String(redisKey, getKey(args)))

	if c.isDebugEnabled() {
		setTraceAttributes(span, args)
	}

	err := c.client.Do(ctx, action)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if c.isDebugEnabled() {
		setDataAttribute(span, rcv)
	}

	return nil
}
