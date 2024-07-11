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
	// TODO change to 1 day
	// length of time to store cached geoip data for
	cacheDuration = time.Hour * 24 * 7
)

type Config interface {
	GeoIPEndpoint() string
	GeoIPAPIKey() string
}

type Location struct {
	config Config
	cache  *cache.Cache
	client *http.Client
}

func New(cfg Config, ca *cache.Cache) *Location {
	return &Location{config: cfg, cache: ca, client: http.DefaultClient}
}

type geoIPResponse struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`

	Error *errResponse `json:"error,omitempty"`
}

type errResponse struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

// TODO trace request

// ByIP performs a geoip query to retrieve a lat and long from a source ip address.
func (l *Location) ByIP(ctx context.Context, sourceIp string) (orb.Point, error) {
	ip := net.ParseIP(sourceIp)
	if ip.IsLoopback() || ip.IsUnspecified() {
		logger.Warn(ctx, "could not parse ip", zap.String("ip", sourceIp))
		sourceIp = "" // use the default.
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s", l.config.GeoIPEndpoint(), sourceIp))
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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
