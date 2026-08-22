package activity

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"ticketgo/internal/httpapi/apierror"
	"ticketgo/internal/httpapi/pagination"
	"ticketgo/pkg/response"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) Handler { return Handler{service: s} }
func (h Handler) Create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	a, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, a)
}
func (h Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	a, err := h.service.ByID(c.Request.Context(), id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, a)
}
func (h Handler) List(c *gin.Context) {
	p, err := pagination.Parse(c)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	a, err := h.service.List(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"activities": a, "limit": p.Limit, "offset": p.Offset})
}
