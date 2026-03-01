package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/boardgov/internal/usecase"
)

type DashboardHandler struct{ uc usecase.DashboardUsecase }

func NewDashboardHandler(uc usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) PendingApprovals(c *gin.Context) {
	items, err := h.uc.PendingApprovals(c.Request.Context(), currentUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) PendingReviews(c *gin.Context) {
	items, err := h.uc.PendingReviews(c.Request.Context(), currentUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) OverdueDocuments(c *gin.Context) {
	items, err := h.uc.OverdueDocuments(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) MeetingStatusSummary(c *gin.Context) {
	items, err := h.uc.MeetingStatusSummary(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) DocumentAgingAnalysis(c *gin.Context) {
	items, err := h.uc.DocumentAgingAnalysis(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) ApprovalTurnaround(c *gin.Context) {
	items, err := h.uc.ApprovalTurnaround(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) RejectionRatios(c *gin.Context) {
	items, err := h.uc.RejectionRatios(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *DashboardHandler) ActivityTimeline(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	items, total, err := h.uc.ActivityTimeline(c.Request.Context(), page, pageSize)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, paginationResponse{Data: items, Page: page, PageSize: pageSize, Total: total})
}

func (h *DashboardHandler) AuditIntegrityStatus(c *gin.Context) {
	status, err := h.uc.AuditIntegrityStatus(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}
