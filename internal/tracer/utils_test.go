package tracer

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestAttribute(t *testing.T) {
	t.Run("it should return the key value for the correct type", func(t *testing.T) {
		kv := Attribute("test", "test")
		require.Equal(t, "STRING", kv.Value.Type().String())
		require.Equal(t, "test", kv.Value.AsString())

		// attribute lib converts int to int64 for some reason, hence this is expected behaviour.
		kv = Attribute("test", 1)
		require.Equal(t, "INT64", kv.Value.Type().String())
		require.Equal(t, int64(1), kv.Value.AsInt64())

		kv = Attribute("test", []string{"test"})
		require.Equal(t, "STRINGSLICE", kv.Value.Type().String())
		require.Equal(t, []string{"test"}, kv.Value.AsStringSlice())

		kv = Attribute("test", true)
		require.Equal(t, "BOOL", kv.Value.Type().String())
		require.Equal(t, true, kv.Value.AsBool())
	})

	t.Run("it should handle pointers correctly", func(t *testing.T) {
		var s = "test"
		kv := Attribute("test", &s)
		require.Equal(t, "STRING", kv.Value.Type().String())
		require.Equal(t, "test", kv.Value.AsString())

		ps := &s
		kv = Attribute("test", ps)
		require.Equal(t, "STRING", kv.Value.Type().String())
		require.Equal(t, "test", kv.Value.AsString())
	})

	t.Run("it should return a zero value for unhandled type", func(t *testing.T) {
		s := http.NoBody
		kv := Attribute("test", s)
		require.Equal(t, "INVALID", kv.Value.Type().String())
		require.Equal(t, "", kv.Value.AsString())
	})
}
