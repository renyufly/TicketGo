import { apiRequest } from "../../api/client";
import type { Order, Page } from "../../api/types";

interface OrderListResponse {
  orders: Order[];
  limit: number;
  offset: number;
}

export async function listOrders(
  token: string,
  limit: number,
  offset: number,
): Promise<Page<Order>> {
  const { data } = await apiRequest<OrderListResponse>(
    `/api/v1/orders?limit=${limit}&offset=${offset}`,
    {
      token,
    },
  );
  return { items: data.orders, limit: data.limit, offset: data.offset };
}

export async function getOrder(token: string, id: string) {
  return (await apiRequest<Order>(`/api/v1/orders/${id}`, { token })).data;
}

export async function cancelOrder(token: string, id: number) {
  return (
    await apiRequest<Order>(`/api/v1/orders/${id}/cancel`, {
      token,
      method: "POST",
    })
  ).data;
}
