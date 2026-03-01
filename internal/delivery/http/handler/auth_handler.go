package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/boardgov/internal/middleware"
	"github.com/yourorg/boardgov/internal/usecase"
	"github.com/yourorg/boardgov/pkg/logger"
)

type AuthHandler struct{ uc usecase.AuthUsecase }

func NewAuthHandler(uc usecase.AuthUsecase) *AuthHandler { return &AuthHandler{uc: uc} }

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type registerRequest struct {
	Email    string `json:"email"     binding:"required,email"`
	Password string `json:"password"  binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	logger.Log.Info().Msg("register endpoint hit")

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Info().Msg("login bind error: " + err.Error())

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password, c.ClientIP())
	if err != nil {
		logger.Log.Info().Msg("login error: " + err.Error())
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.uc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		logger.Log.Info().Msg("refresh error: " + err.Error())
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = h.uc.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id": c.MustGet(middleware.ContextUserID),
		"email":   c.MustGet(middleware.ContextUserEmail),
	})
}
