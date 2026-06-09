package handler

import (
	"net/http"
	"strings"

	"MyRagByCivic/app"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP endpoints for a frontend or API client.
type Handler struct {
	service *app.Service
}

// New creates a handler set backed by the shared app service.
func New(service *app.Service) *Handler {
	return &Handler{service: service}
}

type chatRequest struct {
	Question string `json:"question"`
}

type chatResponse struct {
	Answer string `json:"answer"`
}

type statusResponse struct {
	Ready    bool            `json:"ready"`
	Message  string          `json:"message"`
	Indexing app.IndexReport `json:"indexing"`
}

// Root returns a simple welcome message for browser checks.
func (h *Handler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "RAG API is running",
	})
}

// Health returns the current readiness and last indexing report.
func (h *Handler) Health(c *gin.Context) {
	report := h.service.Status()
	message := "documents are ready"
	if !report.Ready {
		message = "documents are not indexed yet"
	}

	c.JSON(http.StatusOK, statusResponse{
		Ready:    report.Ready,
		Message:  message,
		Indexing: report,
	})
}

// Chat answers one question for the frontend.
func (h *Handler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "send JSON like {\"question\":\"...\"}",
		})
		return
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question cannot be empty",
		})
		return
	}

	answer, err := h.service.Ask(c.Request.Context(), req.Question)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, chatResponse{Answer: answer})
}

// Reindex refreshes the document index without restarting the server.
func (h *Handler) Reindex(c *gin.Context) {
	report, err := h.service.IndexDocumentsFromPDF(c.Request.Context())
	if err != nil && report.ChunksIndexed == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  err.Error(),
			"report": report,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "indexing finished",
		"report":  report,
	})
}
