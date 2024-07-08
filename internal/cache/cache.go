package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mediocregopher/radix/v4"
)

const (
	get   = "GET"
	setEx = "SETEX"
)

const (
	connectionPoolSize = 10
)

var (
	// ErrCacheMiss error is returned when we do not find anything in the cache with the given key.
	ErrCacheMiss = errors.New("cache entry does not exist with the given key")
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
}

// New connect to redis and returns a new instance of the Cache to interact with.
func New(ctx context.Context, password string, host string) (*Cache, error) {
	client, err := (radix.PoolConfig{
		Dialer: radix.Dialer{
			AuthPass: password,
		},
		Size: connectionPoolSize,
	}).New(ctx, "tcp", host)

	if err != nil {
		return nil, err
	}

	return &Cache{client: client}, nil
}

// Client returns the underlying radix redis client.
func (c *Cache) Client() radix.Client {
	return c.client
}

// Close closes the underlying connection to the redis server.
func (c *Cache) Close() error {
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
	action := radix.FlatCmd(rcv, cmd, args...)

	err := c.client.Do(ctx, action)
	if err != nil {
		return err
	}

	return nil
}
