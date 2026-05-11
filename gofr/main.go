// @title Launch Ideas API
// @version 1.0
// @description A small GoFr demo API for managing launch ideas.
// @BasePath /
// @schemes http
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gofr-demo/migrations"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"
)

const (
	baseListIdeasQuery = `
		SELECT id, title, pitch, stage, hype_score, created_at::text, last_boosted_at::text
		FROM launch_ideas
	`
	getIdeaQuery    = baseListIdeasQuery + ` WHERE id = $1`
	createIdeaQuery = `
		INSERT INTO launch_ideas (title, pitch, stage)
		VALUES ($1, $2, $3)
		RETURNING id, title, pitch, stage, hype_score, created_at::text, last_boosted_at::text
	`
	boostIdeaQuery = `
		UPDATE launch_ideas
		SET hype_score = hype_score + 1,
		    last_boosted_at = NOW()
		WHERE id = $1
		RETURNING id, title, pitch, stage, hype_score, created_at::text, last_boosted_at::text
	`
	deleteIdeaQuery = `DELETE FROM launch_ideas WHERE id = $1 RETURNING id`
)

type Idea struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Pitch         string         `json:"pitch"`
	Stage         string         `json:"stage"`
	HypeScore     int            `json:"hypeScore" db:"hype_score"`
	CreatedAt     string         `json:"createdAt" db:"created_at"`
	LastBoostedAt sql.NullString `json:"-" db:"last_boosted_at"`
}

type ideaResponse struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Pitch         string  `json:"pitch"`
	Stage         string  `json:"stage"`
	HypeScore     int     `json:"hypeScore"`
	CreatedAt     string  `json:"createdAt"`
	LastBoostedAt *string `json:"lastBoostedAt,omitempty"`
}

type createIdeaRequest struct {
	Title string `json:"title"`
	Pitch string `json:"pitch"`
	Stage string `json:"stage"`
}

type APIListIdeasData struct {
	Ideas []ideaResponse `json:"ideas"`
	Count int            `json:"count"`
}

type APIListIdeasMetadata struct {
	Filter listIdeasFilter `json:"filter"`
	Host   string          `json:"host"`
}

type APIListIdeasResponse struct {
	Data     APIListIdeasData     `json:"data"`
	Metadata APIListIdeasMetadata `json:"metadata"`
}

type APIIdeaData struct {
	Idea ideaResponse `json:"idea"`
}

type APIIdeaResponse struct {
	Data APIIdeaData `json:"data"`
}

type APIMutationData struct {
	Message string       `json:"message"`
	Idea    ideaResponse `json:"idea"`
}

type APIMutationResponse struct {
	Data APIMutationData `json:"data"`
}

type APIErrorResponse struct {
	Error string `json:"error"`
}

type listIdeasFilter struct {
	Stages  []string `json:"stages,omitempty"`
	Query   string   `json:"query,omitempty"`
	MinHype *int     `json:"minHype,omitempty"`
	Limit   int      `json:"limit"`
}

type httpError struct {
	status  int
	message string
}

func (e httpError) Error() string {
	return e.message
}

func (e httpError) StatusCode() int {
	return e.status
}

func main() {
	app := newApp()
	app.Run()
}

func newApp() *gofr.App {
	app := gofr.New()

	app.Migrate(migrations.All())

	app.GET("/ideas", listIdeas)
	app.GET("/ideas/{id}", getIdea)
	app.POST("/ideas", createIdea)
	app.POST("/ideas/{id}/boost", boostIdea)
	app.DELETE("/ideas/{id}", deleteIdea)

	return app
}

// listIdeas godoc
// @Summary List ideas
// @Description Returns ideas with optional stage, query, minimum hype, and limit filters.
// @Tags ideas
// @Produce json
// @Param stage query string false "Stage filter. Repeat the parameter or pass comma-separated values."
// @Param q query string false "Case-insensitive text search over title and pitch"
// @Param minHype query int false "Minimum hype score"
// @Param limit query int false "Maximum number of ideas to return"
// @Success 200 {object} APIListIdeasResponse
// @Failure 400 {object} APIErrorResponse
// @Router /ideas [get]
func listIdeas(c *gofr.Context) (any, error) {
	defer c.Trace("listIdeas").End()

	filter, err := parseListIdeasFilter(c)
	if err != nil {
		return nil, err
	}

	query, args := buildListIdeasQuery(filter)

	rows, err := c.SQL.QueryContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			c.Errorf("failed to close idea rows: %v", closeErr)
		}
	}()

	ideas := make([]ideaResponse, 0)
	for rows.Next() {
		idea, err := scanIdea(rows)
		if err != nil {
			return nil, err
		}

		ideas = append(ideas, idea)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	c.Infof("listed %d ideas with filter %+v", len(ideas), filter)

	return response.Response{
		Data: map[string]any{
			"ideas": ideas,
			"count": len(ideas),
		},
		Metadata: map[string]any{
			"filter": filter,
			"host":   c.HostName(),
		},
	}, nil
}

// getIdea godoc
// @Summary Get one idea
// @Tags ideas
// @Produce json
// @Param id path int true "Idea ID"
// @Success 200 {object} APIIdeaResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /ideas/{id} [get]
func getIdea(c *gofr.Context) (any, error) {
	defer c.Trace("getIdea").End()

	id, err := ideaID(c)
	if err != nil {
		c.Warnf("invalid idea id: %q", c.PathParam("id"))
		return nil, err
	}

	idea, err := queryIdea(c, getIdeaQuery, id)
	if err != nil {
		c.Errorf("failed to load idea %d: %v", id, err)
		return nil, err
	}

	c.Infof("loaded idea %d", idea.ID)

	return map[string]any{
		"idea": idea,
	}, nil
}

// createIdea godoc
// @Summary Create an idea
// @Tags ideas
// @Accept json
// @Produce json
// @Param request body createIdeaRequest true "Idea payload"
// @Success 201 {object} APIMutationResponse
// @Failure 400 {object} APIErrorResponse
// @Router /ideas [post]
func createIdea(c *gofr.Context) (any, error) {
	defer c.Trace("createIdea").End()

	var req createIdeaRequest
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Pitch = strings.TrimSpace(req.Pitch)
	req.Stage = strings.TrimSpace(req.Stage)

	if req.Title == "" || req.Pitch == "" || req.Stage == "" {
		return nil, badRequest("title, pitch and stage are required")
	}

	idea, err := queryIdea(c, createIdeaQuery, req.Title, req.Pitch, req.Stage)
	if err != nil {
		return nil, err
	}

	c.Infof("created idea %d", idea.ID)

	return map[string]any{
		"message": "idea created",
		"idea":    idea,
	}, nil
}

// boostIdea godoc
// @Summary Boost an idea
// @Tags ideas
// @Produce json
// @Param id path int true "Idea ID"
// @Success 201 {object} APIMutationResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /ideas/{id}/boost [post]
func boostIdea(c *gofr.Context) (any, error) {
	defer c.Trace("boostIdea").End()

	id, err := ideaID(c)
	if err != nil {
		return nil, err
	}

	idea, err := queryIdea(c, boostIdeaQuery, id)
	if err != nil {
		return nil, err
	}

	c.Infof("boosted idea %d to hype score %d", idea.ID, idea.HypeScore)

	return map[string]any{
		"message": "idea boosted",
		"idea":    idea,
	}, nil
}

// deleteIdea godoc
// @Summary Delete an idea
// @Tags ideas
// @Param id path int true "Idea ID"
// @Success 204 "No Content"
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Router /ideas/{id} [delete]
func deleteIdea(c *gofr.Context) (any, error) {
	defer c.Trace("deleteIdea").End()

	id, err := ideaID(c)
	if err != nil {
		return nil, err
	}

	var deletedID int64
	err = c.SQL.QueryRowContext(c, deleteIdeaQuery, id).Scan(&deletedID)
	if err != nil {
		return nil, normalizeSQLError(err)
	}

	c.Infof("deleted idea %d", deletedID)

	return nil, nil
}

func parseListIdeasFilter(c *gofr.Context) (listIdeasFilter, error) {
	filter := listIdeasFilter{
		Stages: cleanStrings(c.Params("stage")),
		Query:  strings.TrimSpace(c.Param("q")),
		Limit:  20,
	}

	if raw := strings.TrimSpace(c.Param("minHype")); raw != "" {
		minHype, err := strconv.Atoi(raw)
		if err != nil || minHype < 0 {
			return listIdeasFilter{}, badRequest("minHype must be a non-negative integer")
		}

		filter.MinHype = &minHype
	}

	if raw := strings.TrimSpace(c.Param("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return listIdeasFilter{}, badRequest("limit must be between 1 and 100")
		}

		filter.Limit = limit
	}

	return filter, nil
}

func buildListIdeasQuery(filter listIdeasFilter) (string, []any) {
	var (
		conditions []string
		stageTerms []string
		args       []any
	)

	for _, stage := range filter.Stages {
		args = append(args, stage)
		stageTerms = append(stageTerms, fmt.Sprintf("stage = $%d", len(args)))
	}

	if len(stageTerms) > 0 {
		conditions = append(conditions, "("+strings.Join(stageTerms, " OR ")+")")
	}

	if filter.Query != "" {
		args = append(args, "%"+filter.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR pitch ILIKE $%d)", len(args), len(args)))
	}

	if filter.MinHype != nil {
		args = append(args, *filter.MinHype)
		conditions = append(conditions, fmt.Sprintf("hype_score >= $%d", len(args)))
	}

	var query strings.Builder
	query.WriteString(baseListIdeasQuery)

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	args = append(args, filter.Limit)
	fmt.Fprintf(&query, " ORDER BY hype_score DESC, created_at DESC LIMIT $%d", len(args))

	return query.String(), args
}

func queryIdea(c *gofr.Context, query string, args ...any) (ideaResponse, error) {
	var idea Idea
	err := c.SQL.QueryRowContext(c, query, args...).Scan(
		&idea.ID,
		&idea.Title,
		&idea.Pitch,
		&idea.Stage,
		&idea.HypeScore,
		&idea.CreatedAt,
		&idea.LastBoostedAt,
	)
	if err != nil {
		return ideaResponse{}, normalizeSQLError(err)
	}

	return toIdeaResponse(idea), nil
}

func cleanStrings(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}

	return cleaned
}

func ideaID(c *gofr.Context) (int64, error) {
	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, badRequest("invalid idea id")
	}

	return id, nil
}

func normalizeSQLError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return httpError{status: http.StatusNotFound, message: "idea not found"}
	}

	return err
}

func badRequest(message string) error {
	return httpError{status: http.StatusBadRequest, message: message}
}

func scanIdea(rows *sql.Rows) (ideaResponse, error) {
	var idea Idea
	err := rows.Scan(
		&idea.ID,
		&idea.Title,
		&idea.Pitch,
		&idea.Stage,
		&idea.HypeScore,
		&idea.CreatedAt,
		&idea.LastBoostedAt,
	)
	if err != nil {
		return ideaResponse{}, err
	}

	return toIdeaResponse(idea), nil
}

func toIdeaResponse(idea Idea) ideaResponse {
	response := ideaResponse{
		ID:        idea.ID,
		Title:     idea.Title,
		Pitch:     idea.Pitch,
		Stage:     idea.Stage,
		HypeScore: idea.HypeScore,
		CreatedAt: idea.CreatedAt,
	}

	if idea.LastBoostedAt.Valid {
		response.LastBoostedAt = &idea.LastBoostedAt.String
	}

	return response
}
