package jwk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

var cache struct {
	m map[string]cacheEntry

	sync.Mutex
}

type cacheEntry struct {
	exp  time.Time
	keys []Jwk
}

func (ce *cacheEntry) Expired() bool {
	return time.Now().Compare(ce.exp) > 0
}

func (ce *cacheEntry) Keys() []Jwk {
	return ce.keys
}

func cacheFetch(ctx context.Context, jwksUri string) ([]Jwk, error) {
	cache.Lock()
	defer cache.Unlock()

	ce, ok := cache.m[jwksUri]
	if ok && !ce.Expired() {
		return ce.Keys(), nil
	}

	keys, err := keysFromUri(ctx, jwksUri)
	if err != nil {
		return nil, err
	}

	if cache.m == nil {
		cache.m = make(map[string]cacheEntry)
	}

	cache.m[jwksUri] = cacheEntry{
		exp:  time.Now().Add(cacheExpMins * time.Minute).Add(time.Duration(rand.Int32N(cacheExpRandMins)) * time.Minute),
		keys: keys,
	}

	return keys, nil
}

func keysFromUri(ctx context.Context, uri string) ([]Jwk, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jwks (%w)", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jwks (%w)", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jwks (%w)", err)
	}

	var jwks Jwks
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to fetch jwks (%w)", err)
	}

	return jwks.Keys, nil
}
