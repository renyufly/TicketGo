import { apiRequest } from "../../api/client";
import type { Item, Page } from "../../api/types";

interface ItemListResponse {
  items: Item[];
  limit: number;
  offset: number;
}

export async function listItems(
  token: string,
  limit: number,
  offset: number,
): Promise<Page<Item>> {
  const { data } = await apiRequest<ItemListResponse>(
    `/api/v1/items?limit=${limit}&offset=${offset}`,
    { token },
  );
  return data;
}

export async function getItem(token: string, id: string) {
  return (await apiRequest<Item>(`/api/v1/items/${id}`, { token })).data;
}

export async function createItem(
  token: string,
  input: {
    name: string;
    description: string;
    price_cents: number;
    status: string;
  },
) {
  return (
    await apiRequest<Item>("/api/v1/items", {
      token,
      method: "POST",
      body: JSON.stringify(input),
    })
  ).data;
}
