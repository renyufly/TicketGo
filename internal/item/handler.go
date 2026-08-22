package item

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
	x, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, x)
}
func (h Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	x, err := h.service.ByID(c.Request.Context(), id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, x)
}
func (h Handler) List(c *gin.Context) {
	page, err := pagination.Parse(c)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	items, err := h.service.List(c.Request.Context(), page.Limit, page.Offset)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"items": items, "limit": page.Limit, "offset": page.Offset})
}
