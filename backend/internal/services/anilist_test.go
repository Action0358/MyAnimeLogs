package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGenresByMALId(t *testing.T) {
	// Mock AniList API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify query contains genres field
		query, ok := requestBody["query"].(string)
		if !ok {
			t.Fatal("Query field not found in request")
		}
		if query == "" {
			t.Fatal("Query is empty")
		}

		// Mock response with genres
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"Media": map[string]interface{}{
					"genres": []string{"Action", "Comedy", "Spy"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create client with mock server endpoint
	client := NewAniListClientWithEndpoint(mockServer.URL)

	// Test GetGenresByMALId
	ctx := context.Background()
	genres, err := client.GetGenresByMALId(ctx, "53065")
	if err != nil {
		t.Fatalf("GetGenresByMALId failed: %v", err)
	}

	// Verify genres are translated to Japanese
	expectedGenres := []string{"アクション", "コメディ", "Spy"} // "Spy" has no translation
	if len(genres) != len(expectedGenres) {
		t.Errorf("Expected %d genres, got %d", len(expectedGenres), len(genres))
	}

	for i, genre := range genres {
		if genre != expectedGenres[i] {
			t.Errorf("Expected genre %s, got %s", expectedGenres[i], genre)
		}
	}
}

func TestTranslateGenre(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Action", "アクション"},
		{"Comedy", "コメディ"},
		{"Drama", "ドラマ"},
		{"Fantasy", "ファンタジー"},
		{"Romance", "ロマンス"},
		{"Sci-Fi", "SF"},
		{"Slice of Life", "日常"},
		{"Unknown Genre", "Unknown Genre"}, // No translation available
	}

	for _, test := range tests {
		result := translateGenre(test.input)
		if result != test.expected {
			t.Errorf("translateGenre(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetGenresByMALId_InvalidMALId(t *testing.T) {
	client := NewAniListClient()
	ctx := context.Background()

	// Test with invalid MAL ID
	_, err := client.GetGenresByMALId(ctx, "invalid")
	if err == nil {
		t.Error("Expected error for invalid MAL ID, got nil")
	}
}

func TestGetGenresByMALId_EmptyMALId(t *testing.T) {
	client := NewAniListClient()
	ctx := context.Background()

	// Test with empty MAL ID
	_, err := client.GetGenresByMALId(ctx, "")
	if err == nil {
		t.Error("Expected error for empty MAL ID, got nil")
	}
}

func TestGetGenresByMALId_NotFound(t *testing.T) {
	// Mock AniList API server returning not found
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return GraphQL error
		response := map[string]interface{}{
			"errors": []map[string]interface{}{
				{
					"message": "Media not found",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewAniListClientWithEndpoint(mockServer.URL)

	// Test GetGenresByMALId with non-existent anime
	ctx := context.Background()
	_, err := client.GetGenresByMALId(ctx, "999999")
	if err == nil {
		t.Error("Expected error for non-existent anime, got nil")
	}
}
