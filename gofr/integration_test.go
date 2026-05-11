package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gofr.dev/pkg/gofr/testutil"
)

func TestIntegrationListIdeas(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker-backed integration test in short mode")
	}

	configs := testutil.NewServerConfigs(t)

	if !waitForPostgres(t) {
		t.Skip("postgres is not available; run docker compose up -d to execute integration tests")
	}

	app := newApp()
	go app.Run()

	t.Cleanup(func() {
		_ = app.Shutdown(context.Background())
	})

	waitForEndpoint(t, configs.HTTPHost+"/ideas")

	resp, err := http.Get(configs.HTTPHost + "/ideas")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Data struct {
			Ideas []ideaResponse `json:"ideas"`
			Count int            `json:"count"`
		} `json:"data"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	assert.GreaterOrEqual(t, payload.Data.Count, 3)
	assert.Len(t, payload.Data.Ideas, payload.Data.Count)
	assert.NotEmpty(t, payload.Data.Ideas[0].Title)
	assert.IsType(t, int64(0), payload.Data.Ideas[0].ID)
}

func waitForPostgres(t *testing.T) bool {
	t.Helper()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_USER", "postgres"),
		envOrDefault("DB_PASSWORD", "postgres"),
		envOrDefault("DB_NAME", "gofr_demo"),
		envOrDefault("DB_SSL_MODE", "disable"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer func() {
		_ = db.Close()
	}()

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return true
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

func waitForEndpoint(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("endpoint %s did not become ready", url)
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
