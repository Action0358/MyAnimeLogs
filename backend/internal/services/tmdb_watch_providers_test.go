package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestTMDBClient creates a TMDBClient pointing at the given test server.
func newTestTMDBClient(serverURL string) *TMDBClient {
	return &TMDBClient{
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
	}
}

// watchProviderResponse builds a TMDBWatchProviders JSON with JP flatrate providers.
func watchProviderResponse(providers []TMDBProvider) TMDBWatchProviders {
	return TMDBWatchProviders{
		ID: 1,
		Results: map[string]TMDBCountryResult{
			"JP": {Flatrate: providers},
		},
	}
}

// emptyWatchProviderResponse returns a response with no JP providers.
func emptyWatchProviderResponse() TMDBWatchProviders {
	return TMDBWatchProviders{ID: 1, Results: map[string]TMDBCountryResult{}}
}

func TestGetJapaneseWatchProvidersWithFallback_TVFound(t *testing.T) {
	netflix := []TMDBProvider{{ProviderID: 8, ProviderName: "Netflix", LogoPath: "/netflix.png"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tv/") {
			json.NewEncoder(w).Encode(watchProviderResponse(netflix))
			return
		}
		// movie endpoint should not be called
		t.Error("unexpected call to movie endpoint")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestTMDBClient(server.URL)
	// Override the base URL by replacing httpClient with a transport that rewrites URLs
	client.httpClient = &http.Client{Transport: &rewriteTransport{base: server.URL}}

	providers, err := client.GetJapaneseWatchProvidersWithFallback(context.Background(), 1575337, "tv")
	assert.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "Netflix", providers[0].ProviderName)
}

func TestGetJapaneseWatchProvidersWithFallback_TVFailsFallbackToMovie(t *testing.T) {
	netflix := []TMDBProvider{{ProviderID: 8, ProviderName: "Netflix", LogoPath: "/netflix.png"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tv/") {
			// TV returns empty (no JP providers)
			json.NewEncoder(w).Encode(emptyWatchProviderResponse())
			return
		}
		if strings.Contains(r.URL.Path, "/movie/") {
			json.NewEncoder(w).Encode(watchProviderResponse(netflix))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestTMDBClient(server.URL)
	client.httpClient = &http.Client{Transport: &rewriteTransport{base: server.URL}}

	providers, err := client.GetJapaneseWatchProvidersWithFallback(context.Background(), 1575337, "tv")
	assert.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "Netflix", providers[0].ProviderName)
}

func TestGetJapaneseWatchProvidersWithFallback_MovieFailsFallbackToTV(t *testing.T) {
	crunchyroll := []TMDBProvider{{ProviderID: 283, ProviderName: "Crunchyroll", LogoPath: "/cr.png"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/movie/") {
			json.NewEncoder(w).Encode(emptyWatchProviderResponse())
			return
		}
		if strings.Contains(r.URL.Path, "/tv/") {
			json.NewEncoder(w).Encode(watchProviderResponse(crunchyroll))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestTMDBClient(server.URL)
	client.httpClient = &http.Client{Transport: &rewriteTransport{base: server.URL}}

	providers, err := client.GetJapaneseWatchProvidersWithFallback(context.Background(), 999, "movie")
	assert.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "Crunchyroll", providers[0].ProviderName)
}

func TestGetJapaneseWatchProvidersWithFallback_BothEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(emptyWatchProviderResponse())
	}))
	defer server.Close()

	client := newTestTMDBClient(server.URL)
	client.httpClient = &http.Client{Transport: &rewriteTransport{base: server.URL}}

	providers, err := client.GetJapaneseWatchProvidersWithFallback(context.Background(), 999, "tv")
	assert.NoError(t, err)
	assert.Nil(t, providers)
}

func TestGetJapaneseWatchProvidersWithFallback_APIErrorPreserved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newTestTMDBClient(server.URL)
	client.httpClient = &http.Client{Transport: &rewriteTransport{base: server.URL}}

	providers, err := client.GetJapaneseWatchProvidersWithFallback(context.Background(), 999, "tv")
	assert.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "429")
}

// rewriteTransport rewrites TMDB API URLs to point at the test server.
type rewriteTransport struct {
	base string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace tmdbBaseURL with test server URL
	newURL := strings.Replace(req.URL.String(), tmdbBaseURL, t.base, 1)
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(newReq)
}
