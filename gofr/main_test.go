package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"
	gofrHTTP "gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/logging"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("GOFR_TELEMETRY", "false"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestCreateIdea(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	ctx := newTestContext(t, mockContainer, testRequest{
		method: http.MethodPost,
		path:   "/ideas",
		body:   `{"title":"Latency radar","pitch":"Explains slow endpoints before a customer notices.","stage":"beta"}`,
	})

	mocks.SQL.ExpectQuery(createIdeaQuery).
		WithArgs("Latency radar", "Explains slow endpoints before a customer notices.", "beta").
		WillReturnRows(mocks.SQL.NewRows([]string{"id", "title", "pitch", "stage", "hype_score", "created_at", "last_boosted_at"}).
			AddRow(int64(42), "Latency radar", "Explains slow endpoints before a customer notices.", "beta", 0, "2026-03-15 10:00:00+00", nil))

	resp, err := createIdea(ctx)
	require.NoError(t, err)

	payload := resp.(map[string]any)
	assert.Equal(t, "idea created", payload["message"])

	idea := payload["idea"].(ideaResponse)
	assert.Equal(t, int64(42), idea.ID)
	assert.Equal(t, "Latency radar", idea.Title)
	assert.Nil(t, idea.LastBoostedAt)
}

func TestBoostIdeaRejectsInvalidInt64(t *testing.T) {
	ctx := newTestContext(t, &container.Container{Logger: logging.NewMockLogger(logging.INFO)}, testRequest{
		method:     http.MethodPost,
		path:       "/ideas/not-a-number/boost",
		pathParams: map[string]string{"id": "not-a-number"},
	})

	resp, err := boostIdea(ctx)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid idea id")
}

func TestBuildListIdeasQueryUsesFiltersAndLimit(t *testing.T) {
	minHype := 3
	query, args := buildListIdeasQuery(listIdeasFilter{
		Stages:  []string{"prototype", "beta"},
		Query:   "release",
		MinHype: &minHype,
		Limit:   10,
	})

	assert.Contains(t, compactSQL(query), "WHERE (stage = $1 OR stage = $2) AND (title ILIKE $3 OR pitch ILIKE $3) AND hype_score >= $4 ORDER BY hype_score DESC, created_at DESC LIMIT $5")
	assert.Equal(t, []any{"prototype", "beta", "%release%", 3, 10}, args)
}

func TestParseListIdeasFilterRejectsInvalidLimit(t *testing.T) {
	ctx := newTestContext(t, &container.Container{Logger: logging.NewMockLogger(logging.INFO)}, testRequest{
		method: http.MethodGet,
		path:   "/ideas?limit=bad",
	})

	_, err := parseListIdeasFilter(ctx)
	require.Error(t, err)

	var statusErr interface{ StatusCode() int }
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusBadRequest, statusErr.StatusCode())
}

func compactSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

type testRequest struct {
	method     string
	path       string
	body       string
	pathParams map[string]string
}

func (r testRequest) httpRequest() *http.Request {
	var body io.Reader = http.NoBody
	if r.body != "" {
		body = bytes.NewBufferString(r.body)
	}

	req := httptest.NewRequest(r.method, r.path, body)
	if r.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if len(r.pathParams) > 0 {
		req = mux.SetURLVars(req, r.pathParams)
	}

	return req
}

func newTestContext(t *testing.T, cont *container.Container, req testRequest) *gofr.Context {
	t.Helper()

	httpReq := req.httpRequest()
	ctx := httpReq.Context()

	return &gofr.Context{
		Context:       ctx,
		Request:       gofrHTTP.NewRequest(httpReq),
		Container:     cont,
		ContextLogger: *logging.NewContextLogger(ctx, cont.Logger),
	}
}
