// Order订单模块的 HTTP Handler（Controller控制器层）
// 负责接收 Gin 的 HTTP 请求、解析参数，然后调用 Service 处理真正的业务逻辑
// HTTP 请求 → Handler 解析参数/用户 → Service 处理业务 → 返回 JSON

/* 秒杀（Seckill）：典型的高并发抢购场景
商品数量很少，但同一时间有大量用户抢购，先抢到的人才能下单.
注意：防止超卖等问题.
*/
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

// Handler 自己不实现秒杀、取消订单等业务，而是交给h.service
// Handler → Service 分层
type Handler struct{ service *Service }

func NewHandler(s *Service) Handler { return Handler{service: s} }

// 秒杀下单： POST /activities/123/seckill
func (h Handler) Seckill(c *gin.Context) {
	// 从 URL 获取活动 ID， 把字符串 "123" 转成 int64
	activityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	var in SeckillInput
	// 解析请求 JSON, 如 "quantity": 2
	if err = c.ShouldBindJSON(&in); err != nil {
		apierror.WriteValidation(c, err)
		return
	}

	// 从 Gin Context 获取之前认证中间件保存的 JWT 用户信息
	claims, _ := auth.ClaimsFrom(c)

	//调用 Service 执行真正的秒杀逻辑
	o, err := h.service.Seckill(c.Request.Context(), claims.UserID, activityID, in.Quantity)
	if err != nil {
		apierror.Write(c, err)
		return
	}

	// 成功后返回 HTTP 201 + 创建的订单
	response.JSON(c, http.StatusCreated, o)
}

// 查询单个订单： GET /orders/123
func (h Handler) Get(c *gin.Context) {
	// 先把 123 转成订单 ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	// 获取当前登录用户
	claims, _ := auth.ClaimsFrom(c)

	// 调用 service 查询该用户的订单
	o, err := h.service.ByID(c.Request.Context(), claims.UserID, id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, o)
}

// 查询订单列表：GET /orders?limit=20&offset=40
func (h Handler) List(c *gin.Context) {
	// 解析分页参数
	p, err := pagination.Parse(c)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}
	// // 获取当前登录用户
	claims, _ := auth.ClaimsFrom(c)

	// 调用service层 查询当前用户的订单
	orders, err := h.service.List(c.Request.Context(), claims.UserID, p.Limit, p.Offset)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"orders": orders, "limit": p.Limit, "offset": p.Offset})
}

// 取消订单：POST /orders/123/cancel
func (h Handler) Cancel(c *gin.Context) {
	// 获取订单 ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierror.WriteValidation(c, err)
		return
	}

	// // 获取当前登录用户
	claims, _ := auth.ClaimsFrom(c)
	// 让 Service 执行取消订单逻辑
	o, err := h.service.Cancel(c.Request.Context(), claims.UserID, id)
	if err != nil {
		apierror.Write(c, err)
		return
	}
	response.JSON(c, http.StatusOK, o)
}
