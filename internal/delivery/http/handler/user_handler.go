package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/boardgov/internal/domain"
)

type UserHandler struct{ repo domain.UserRepository }

func NewUserHandler(repo domain.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) List(c *gin.Context) {
	page, _     := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	users, total, err := h.repo.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if users == nil {
		users = []*domain.User{}
	}
	c.JSON(http.StatusOK, paginationResponse{Data: users, Page: page, PageSize: pageSize, Total: total})
}
