package location

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/nickbadlose/muzz/internal/cache"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/paulmach/orb"
	"go.uber.org/zap"
)

const (
	// length of time to cache the geoip response for.
	cacheDuration = time.Hour * 24
)

// Config is the interface to retrieve configuration secrets from the environment.
type Config interface {
	// GeoIPEndpoint retrieves the geo IP endpoint.
	GeoIPEndpoint() string
	// GeoIPAPIKey retrieves the geo IP API key.
	GeoIPAPIKey() string
}

// Location is the service which handles all location based queries.
type Location struct {
	config Config
	cache  *cache.Cache
	client *http.Client
}

// New builds a new *Location.
func New(cfg Config, ca *cache.Cache) (*Location, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if ca == nil {
		return nil, errors.New("cache cannot be nil")
	}
	return &Location{config: cfg, cache: ca, client: http.DefaultClient}, nil
}

// the expected response from the geo IP service.
type geoIPResponse struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`

	Error *errResponse `json:"error,omitempty"`
}

// the expected error response from the geo IP service.
type errResponse struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

// ByIP performs a geoip query to retrieve latitude and longitude coordinates from a source ip address.
//
// Responses are cached in redis for 1 since IP location data rarely changes using the endpoint as the key.
func (l *Location) ByIP(ctx context.Context, sourceIP string) (orb.Point, error) {
	ip := net.ParseIP(sourceIP)
	if ip.IsLoopback() || ip.IsUnspecified() {
		logger.Warn(ctx, "could not parse ip", zap.String("ip", sourceIP))
		sourceIP = "" // use the default.
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s", l.config.GeoIPEndpoint(), sourceIP))
	if err != nil {
		logger.Error(ctx, "parsing url from env", err)
		return orb.Point{}, err
	}

	v := &url.Values{
		"access_key": []string{l.config.GeoIPAPIKey()},
		"output":     []string{"json"},
	}
	u.RawQuery = v.Encode()

	endpoint := u.String()

	gRes := new(geoIPResponse)
	err = l.cache.Get(ctx, endpoint, gRes)
	if err == nil {
		return orb.Point{gRes.Lon, gRes.Lat}, nil
	}
	if !errors.Is(err, cache.ErrCacheMiss) {
		logger.Error(ctx, "getting geoip data from cache", err, zap.String("key", endpoint))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		logger.Error(ctx, "building new geoip request", err)
		return orb.Point{}, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := l.client.Do(req)
	if err != nil {
		logger.Error(ctx, "performing geoip request", err)
		return orb.Point{}, err
	}
	defer func() {
		logger.MaybeError(ctx, "closing geoip response body", res.Body.Close())
	}()

	err = json.NewDecoder(res.Body).Decode(&gRes)
	if err != nil {
		logger.Error(ctx, "decoding geoip response", err)
		return orb.Point{}, err
	}

	if gRes.Error != nil {
		logger.Error(
			ctx,
			"response from geoip request",
			errors.New(gRes.Error.Info),
			zap.String("type", gRes.Error.Type),
			zap.Int("code", gRes.Error.Code),
		)
		return orb.Point{}, errors.New(gRes.Error.Info)
	}

	err = l.cache.SetEx(ctx, endpoint, gRes, cacheDuration)
	if err != nil {
		logger.Warn(ctx, "setting geoip data in cache", zap.String("key", endpoint), zap.Error(err))
	}

	return orb.Point{gRes.Lon, gRes.Lat}, nil
}
