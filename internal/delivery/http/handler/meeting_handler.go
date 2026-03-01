package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/internal/usecase"
)

type MeetingHandler struct{ uc usecase.MeetingUsecase }

func NewMeetingHandler(uc usecase.MeetingUsecase) *MeetingHandler {
	return &MeetingHandler{uc: uc}
}

type createMeetingRequest struct {
	Title       string `json:"title"        binding:"required"`
	MeetingType string `json:"meeting_type" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
	Location    string `json:"location"`
	Agenda      string `json:"agenda"`
}

func (h *MeetingHandler) Create(c *gin.Context) {
	var req createMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheduledAt, err := parseTime(req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduled_at, use RFC3339"})
		return
	}

	m := &domain.Meeting{
		Title:       req.Title,
		MeetingType: req.MeetingType,
		ScheduledAt: scheduledAt,
		Location:    req.Location,
		Agenda:      req.Agenda,
		CreatedBy:   currentUserID(c),
	}

	if err := h.uc.Create(c.Request.Context(), m); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *MeetingHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MeetingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := domain.MeetingFilter{
		Status:   domain.MeetingStatus(c.Query("status")),
		Page:     page,
		PageSize: pageSize,
	}

	meetings, total, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, paginationResponse{Data: meetings, Page: page, PageSize: pageSize, Total: total})
}
