package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

// setupCache helper function to start the mocked cache service and
// returns a function which cleans up.
func setupCache(t *testing.T) (*Cache, *miniredis.Miniredis, func()) {
	srv := miniredis.RunT(t)
	cache, err := New(context.TODO(), &Credentials{Host: srv.Addr()})
	require.NoError(t, err)

	return cache, srv, func() {
		require.NoError(t, cache.Close())
	}
}

func TestNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cache testing in short mode")
	}

	t.Run("it should create a new cache without errors, configured correctly", func(t *testing.T) {
		srv := miniredis.RunT(t)
		cache, err := New(context.TODO(), &Credentials{Host: srv.Addr()})
		require.NoError(t, err)
		require.NotNil(t, cache)

		client, err := cache.Client()
		require.NoError(t, err)
		require.NotNil(t, client)

		require.NotNil(t, cache.config)
		require.False(t, cache.config.debugEnabled)
		require.Equal(t, connectionPoolSize, cache.config.connectionPoolSize)
		require.NotNil(t, cache.config.tracerProvider)

		require.NoError(t, cache.Close())
	})

	t.Run("it should configure the cache config correctly", func(t *testing.T) {
		srv := miniredis.RunT(t)
		cache, err := New(
			context.TODO(),
			&Credentials{Host: srv.Addr()},
			WithDebugMode(true),
			WithConnectionPoolSize(1),
			WithTraceProvider(nil),
		)
		require.NoError(t, err)
		require.NotNil(t, cache)
		require.NotNil(t, cache.config)
		require.True(t, cache.config.debugEnabled)
		require.Equal(t, 1, cache.config.connectionPoolSize)
		require.Nil(t, cache.config.tracerProvider)
	})

	t.Run("error: nil credentials", func(t *testing.T) {
		cache, err := New(context.TODO(), nil)
		require.Nil(t, cache)
		require.Error(t, err)
		require.Equal(t, "credentials must be provided", err.Error())
	})
}

func TestCacheGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cache testing in short mode")
	}

	cache, srv, cleanup := setupCache(t)
	t.Cleanup(cleanup)

	t.Run("data does not exist in cache", func(t *testing.T) {
		var resp string
		err := cache.Get(context.TODO(), "testKey", &resp)

		require.Error(t, err)
		require.Equal(t, ErrCacheMiss, err)
		require.NotEqual(t, "testValue", resp)
	})

	t.Run("data exists in cache", func(t *testing.T) {
		var resp string

		data, err := json.Marshal("testValue")
		require.NoError(t, err)

		err = srv.Set("testKey", string(data))
		require.NoError(t, err)

		err = cache.Get(context.TODO(), "testKey", &resp)

		require.NoError(t, err)
		require.Equal(t, "testValue", resp)
	})

	t.Run("redis returns error", func(t *testing.T) {
		srv.SetError("redisErrorMessage")

		var resp string
		err := cache.Get(context.TODO(), "testKey", &resp)

		require.Error(t, err)
		require.Equal(t, "response returned from Conn: redisErrorMessage", err.Error())
		require.NotEqual(t, "testValue", resp)
	})
}

func TestCacheSetEx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cache testing in short mode")
	}

	cache, srv, cleanup := setupCache(t)
	t.Cleanup(cleanup)

	expiration := time.Second * 5

	t.Run("redis returns no error", func(t *testing.T) {
		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)
		require.NoError(t, err)
	})

	t.Run("expiration works as expected", func(t *testing.T) {
		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)
		require.NoError(t, err)

		srv.FastForward(time.Second * 10)

		val, err := srv.Get("testKey")
		require.Error(t, err)
		require.Equal(t, "ERR no such key", err.Error())
		require.NotEqual(t, "testValue", val)
	})

	t.Run("redis returns error", func(t *testing.T) {
		srv.SetError("redisErrorMessage")

		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)

		require.Error(t, err)
		require.Equal(t, "response returned from Conn: redisErrorMessage", err.Error())
	})

	t.Run("json marshalling fails", func(t *testing.T) {
		// invalid type that can't be marshalled into json.
		x := map[string]interface{}{
			"foo": make(chan int),
		}

		err := cache.SetEx(context.TODO(), "testKey", x, expiration)
		require.Error(t, err)
	})
}
