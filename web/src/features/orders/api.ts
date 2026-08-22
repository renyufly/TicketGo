// 前端调用后端的订单 API，
// 包括查询订单列表、查询单个订单、取消订单

/*
apiRequest：项目封装好的 HTTP 请求函数，可以理解成封装后的 fetch。
Order：订单的数据类型。
Page：分页数据类型。
*/
import { apiRequest } from "../../api/client";
import type { Order, Page } from "../../api/types";

/* 定义后端返回格式：
{
  "orders": [...],
  "limit": 10,
  "offset": 0
}
*/
interface OrderListResponse {
  orders: Order[];
  limit: number;
  offset: number;
}

// 获取订单列表
/*
token：JWT 登录凭证
limit：一次查多少条
offset：从第几条开始查
*/
export async function listOrders(
  token: string,
  limit: number,
  offset: number,
): Promise<Page<Order>> {
  /* 请求
  GET /api/v1/orders?limit=10&offset=0
  Authorization: Bearer xxx
  */
  const { data } = await apiRequest<OrderListResponse>(
    `/api/v1/orders?limit=${limit}&offset=${offset}`,
    {
      token,
    },
  );

  // 把后端的 orders 转换成前端统一的分页结构 Page<Order>
  return { items: data.orders, limit: data.limit, offset: data.offset };
}

// 获取单个订单
export async function getOrder(token: string, id: string) {
  return (await apiRequest<Order>(`/api/v1/orders/${id}`, { token })).data;
}

// 取消订单
export async function cancelOrder(token: string, id: number) {
  return (
    await apiRequest<Order>(`/api/v1/orders/${id}/cancel`, {
      token,
      method: "POST",
    })
  ).data;
}
