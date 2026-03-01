package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/internal/usecase"
)

type DocumentHandler struct{ uc usecase.DocumentUsecase }

func NewDocumentHandler(uc usecase.DocumentUsecase) *DocumentHandler {
	return &DocumentHandler{uc: uc}
}

type createDocumentRequest struct {
	MeetingID    string `json:"meeting_id"    binding:"required"`
	Title        string `json:"title"         binding:"required"`
	Description  string `json:"description"`
	WorkflowMode string `json:"workflow_mode"`
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var req createDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mID, err := uuid.Parse(req.MeetingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting_id"})
		return
	}

	wm := domain.WorkflowMode(req.WorkflowMode)
	if wm == "" {
		wm = domain.WorkflowSequential
	}

	doc := &domain.Document{
		MeetingID:    mID,
		Title:        req.Title,
		Description:  req.Description,
		WorkflowMode: wm,
		CreatedBy:    currentUserID(c),
	}

	if err := h.uc.Create(c.Request.Context(), doc); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, doc)
}

// ListByMeeting returns all documents for a meeting as a plain array.
// Called by GET /meetings/:id/documents — the frontend expects Document[] not a paginated wrapper.
func (h *DocumentHandler) ListByMeeting(c *gin.Context) {
	meetingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meeting id"})
		return
	}
	filter := domain.DocumentFilter{
		MeetingID: &meetingID,
		Page:      1,
		PageSize:  200,
	}
	docs, _, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if docs == nil {
		docs = []*domain.Document{}
	}
	c.JSON(http.StatusOK, docs)
}

func (h *DocumentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	doc, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := domain.DocumentFilter{
		Status:   domain.DocumentStatus(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}
	if mid := c.Query("meeting_id"); mid != "" {
		id, err := uuid.Parse(mid)
		if err == nil {
			filter.MeetingID = &id
		}
	}

	docs, total, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, paginationResponse{Data: docs, Page: page, PageSize: pageSize, Total: total})
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 30*1024*1024))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	version, err := h.uc.UploadVersion(c.Request.Context(), docID, currentUserID(c),
		header.Filename, mimeType, data, c.ClientIP())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, version)
}

func (h *DocumentHandler) Download(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	versionNum, _ := strconv.Atoi(c.DefaultQuery("version", "0"))

	data, version, err := h.uc.DownloadVersion(c.Request.Context(), docID, versionNum, currentUserID(c), c.ClientIP())
	if err != nil {
		respondError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, version.FileName))
	c.Header("Content-Type", version.MimeType)
	c.Header("X-SHA256", version.SHA256Hash)
	c.Data(http.StatusOK, version.MimeType, data)
}

func (h *DocumentHandler) GetReviewers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	reviewers, err := h.uc.GetReviewers(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, reviewers)
}

func (h *DocumentHandler) GetApprovers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	approvers, err := h.uc.GetApprovers(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, approvers)
}

func (h *DocumentHandler) SetReviewers(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDs, err := parseUUIDs(req.UserIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.SetReviewers(c.Request.Context(), docID, userIDs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reviewers updated"})
}

func (h *DocumentHandler) SetApprovers(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDs, err := parseUUIDs(req.UserIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.SetApprovers(c.Request.Context(), docID, userIDs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "approvers updated"})
}

func (h *DocumentHandler) Submit(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Submit(c.Request.Context(), docID, currentUserID(c), c.ClientIP()); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "submitted"})
}

func (h *DocumentHandler) Review(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Action   string `json:"action"   binding:"required,oneof=approved rejected"`
		Comments string `json:"comments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.Review(c.Request.Context(), docID, currentUserID(c), req.Action, req.Comments, c.ClientIP()); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "review recorded"})
}

func (h *DocumentHandler) Approve(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Action   string `json:"action"   binding:"required,oneof=approved rejected"`
		Comments string `json:"comments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.Approve(c.Request.Context(), docID, currentUserID(c), req.Action, req.Comments, c.ClientIP()); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "approval recorded"})
}

func (h *DocumentHandler) StartReview(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.StartReview(c.Request.Context(), docID, currentUserID(c), c.ClientIP()); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "review started"})
}

func (h *DocumentHandler) Archive(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Archive(c.Request.Context(), docID, currentUserID(c), c.ClientIP()); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "archived"})
}

func (h *DocumentHandler) ListVersions(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	versions, err := h.uc.ListVersions(c.Request.Context(), docID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, versions)
}

// helpers

func parseUUIDs(ss []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(ss))
	for i, s := range ss {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid %q", s)
		}
		ids[i] = id
	}
	return ids, nil
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
