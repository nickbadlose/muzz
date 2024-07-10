package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// setupCache helper function to start the mocked cache service and
// returns a function which cleans up.
func setupCache(t *testing.T) (*Cache, *miniredis.Miniredis, func()) {
	srv := miniredis.RunT(t)
	cache, err := New(context.TODO(), "", srv.Addr())
	assert.NoError(t, err)

	return cache, srv, func() {
		assert.NoError(t, cache.Close())
	}
}

func TestCacheGet(t *testing.T) {
	cache, srv, cleanup := setupCache(t)
	t.Cleanup(cleanup)

	t.Run("data does not exist in cache", func(t *testing.T) {
		var resp string
		err := cache.Get(context.TODO(), "testKey", &resp)

		assert.Error(t, err)
		assert.Equal(t, ErrCacheMiss, err)
		assert.NotEqual(t, "testValue", resp)
	})

	t.Run("data exists in cache", func(t *testing.T) {
		var resp string

		data, err := json.Marshal("testValue")
		assert.NoError(t, err)

		err = srv.Set("testKey", string(data))
		assert.NoError(t, err)

		err = cache.Get(context.TODO(), "testKey", &resp)

		assert.NoError(t, err)
		assert.Equal(t, "testValue", resp)
	})

	t.Run("redis returns error", func(t *testing.T) {
		srv.SetError("redisErrorMessage")

		var resp string
		err := cache.Get(context.TODO(), "testKey", &resp)

		assert.Error(t, err)
		assert.Equal(t, "response returned from Conn: redisErrorMessage", err.Error())
		assert.NotEqual(t, "testValue", resp)
	})
}

func TestCacheSetEx(t *testing.T) {
	cache, srv, cleanup := setupCache(t)
	t.Cleanup(cleanup)

	expiration := time.Second * 5

	t.Run("redis returns no error", func(t *testing.T) {
		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)
		assert.NoError(t, err)
	})

	t.Run("expiration works as expected", func(t *testing.T) {
		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)
		assert.NoError(t, err)

		srv.FastForward(time.Second * 10)

		val, err := srv.Get("testKey")
		assert.Error(t, err)
		assert.Equal(t, "ERR no such key", err.Error())
		assert.NotEqual(t, "testValue", val)
	})

	t.Run("redis returns error", func(t *testing.T) {
		srv.SetError("redisErrorMessage")

		err := cache.SetEx(context.TODO(), "testKey", "testValue", expiration)

		assert.Error(t, err)
		assert.Equal(t, "response returned from Conn: redisErrorMessage", err.Error())
	})

	t.Run("json marshalling fails", func(t *testing.T) {
		// invalid type that can't be marshalled into json.
		x := map[string]interface{}{
			"foo": make(chan int),
		}

		err := cache.SetEx(context.TODO(), "testKey", x, expiration)
		assert.Error(t, err)
	})
}
