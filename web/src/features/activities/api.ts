import { apiRequest } from "../../api/client";
import type { Activity, Order, Page } from "../../api/types";

interface ActivityListResponse {
  activities: Activity[];
  limit: number;
  offset: number;
}

export async function listActivities(
  token: string,
  limit: number,
  offset: number,
): Promise<Page<Activity>> {
  const { data } = await apiRequest<ActivityListResponse>(
    `/api/v1/activities?limit=${limit}&offset=${offset}`,
    { token },
  );
  return { items: data.activities, limit: data.limit, offset: data.offset };
}

export async function getActivity(token: string, id: string) {
  return (await apiRequest<Activity>(`/api/v1/activities/${id}`, { token }))
    .data;
}

export async function createActivity(
  token: string,
  input: {
    item_id: number;
    name: string;
    price_cents: number;
    starts_at: string;
    ends_at: string;
    status: string;
    total: number;
  },
) {
  return (
    await apiRequest<Activity>("/api/v1/activities", {
      token,
      method: "POST",
      body: JSON.stringify(input),
    })
  ).data;
}

export async function seckill(
  token: string,
  activityId: string,
  quantity: number,
) {
  return (
    await apiRequest<Order>(`/api/v1/activities/${activityId}/seckill`, {
      token,
      method: "POST",
      body: JSON.stringify({ quantity }),
    })
  ).data;
}
