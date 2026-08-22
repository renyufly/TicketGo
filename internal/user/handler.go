package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ticketgo/internal/auth"
	"ticketgo/internal/httpapi/apierror"
	"ticketgo/pkg/response"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) Handler { return Handler{service: service} }
func (h Handler) Create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	u, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, u)
}
func (h Handler) Login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	token, err := h.service.Login(c.Request.Context(), in)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"access_token": token, "token_type": "Bearer"})
}
func (h Handler) Me(c *gin.Context) {
	claims, _ := auth.ClaimsFrom(c)
	u, err := h.service.ByID(c.Request.Context(), claims.UserID)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, u)
}
