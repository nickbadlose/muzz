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

func TestAttributesFromRequestDump(t *testing.T) {
	var (
		testLoginRequest      = []byte("POST /login HTTP/1.1\nTrue-Client-Ip: 51.146.90.158\nContent-Type: application/json\nUser-Agent: PostmanRuntime/7.39.0\nAccept: */*\nPostman-Token: c1112fd8-1fe9-4d62-b2bb-84f45be5e2dc\nHost: localhost:3000\nAccept-Encoding: gzip, deflate, br\nConnection: keep-alive\nContent-Length: 62\r\n\r\n{\n    \"email\": \"test1@test.com\",\n    \"password\": \"Pa55w0rd!\"\n}")
		testCreateUserRequest = []byte("POST /user/create HTTP/1.1\nHost: localhost:3000\nAccept: */*\nAccept-Encoding: gzip, deflate, br\nConnection: keep-alive\nContent-Length: 207\nContent-Type: application/json\nPostman-Token: bfcfe179-9277-4f61-8fbd-436a27f125c6\nTrue-Client-Ip: 51.146.90.158\nUser-Agent: PostmanRuntime/7.39.0\r\n\r\n{\n    \"email\": \"test16@test.com\",\n    \"password\": \"Pa55w0rd!\",\n    \"name\": \"test16\",\n    \"gender\": \"male\",\n    \"age\": 29,\n    \"location\": {\n        \"latitude\": 53.4808,\n        \"longitude\": -2.244644\n    }\n}")
	)

	t.Run("should split body and headers into two separate strings", func(t *testing.T) {
		attr := attributesFromRequestDump(testLoginRequest)
		require.Equal(t, 2, len(attr))
		require.Equal(
			t,
			"POST /login HTTP/1.1\nTrue-Client-Ip: 51.146.90.158\nContent-Type: application/json\nUser-Agent: PostmanRuntime/7.39.0\nAccept: */*\nPostman-Token: c1112fd8-1fe9-4d62-b2bb-84f45be5e2dc\nHost: localhost:3000\nAccept-Encoding: gzip, deflate, br\nConnection: keep-alive\nContent-Length: 62",
			attr[0].Value.AsString(),
		)
		require.Equal(
			t,
			"{\n    \"email\": \"test1@test.com\",\n    \"password\": \"Pa55w0rd!\"\n}",
			attr[1].Value.AsString(),
		)

		attr = attributesFromRequestDump(testCreateUserRequest)
		require.Equal(t, 2, len(attr))
		require.Equal(
			t,
			"POST /user/create HTTP/1.1\nHost: localhost:3000\nAccept: */*\nAccept-Encoding: gzip, deflate, br\nConnection: keep-alive\nContent-Length: 207\nContent-Type: application/json\nPostman-Token: bfcfe179-9277-4f61-8fbd-436a27f125c6\nTrue-Client-Ip: 51.146.90.158\nUser-Agent: PostmanRuntime/7.39.0",
			attr[0].Value.AsString(),
		)
		require.Equal(
			t,
			"{\n    \"email\": \"test16@test.com\",\n    \"password\": \"Pa55w0rd!\",\n    \"name\": \"test16\",\n    \"gender\": \"male\",\n    \"age\": 29,\n    \"location\": {\n        \"latitude\": 53.4808,\n        \"longitude\": -2.244644\n    }\n}",
			attr[1].Value.AsString(),
		)
	})

	t.Run("should not split the data when supplied a dump that has no separator matched.", func(t *testing.T) {
		attr := attributesFromRequestDump([]byte("data with no separator"))
		require.Equal(t, 1, len(attr))
		require.Equal(
			t,
			"data with no separator",
			attr[0].Value.AsString(),
		)
	})
}
