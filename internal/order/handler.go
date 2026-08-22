package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"ticketgo/internal/auth"
	"ticketgo/internal/httpapi/apierror"
	"ticketgo/internal/httpapi/pagination"
	"ticketgo/pkg/response"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) Handler { return Handler{service: s} }
func (h Handler) Seckill(c *gin.Context) {
	activityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	var in SeckillInput
	if err = c.ShouldBindJSON(&in); err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	claims, _ := auth.ClaimsFrom(c)
	o, err := h.service.Seckill(c.Request.Context(), claims.UserID, activityID, in.Quantity)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, o)
}
func (h Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	claims, _ := auth.ClaimsFrom(c)
	o, err := h.service.ByID(c.Request.Context(), claims.UserID, id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, o)
}
func (h Handler) List(c *gin.Context) {
	p, err := pagination.Parse(c)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	claims, _ := auth.ClaimsFrom(c)
	orders, err := h.service.List(c.Request.Context(), claims.UserID, p.Limit, p.Offset)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"orders": orders, "limit": p.Limit, "offset": p.Offset})
}
func (h Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	claims, _ := auth.ClaimsFrom(c)
	o, err := h.service.Cancel(c.Request.Context(), claims.UserID, id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, o)
}
