export type UserRole = "customer" | "admin";

export interface User {
  id: number;
  email: string;
  role: UserRole;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Item {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Activity {
  id: number;
  item_id: number;
  name: string;
  price_cents: number;
  starts_at: string;
  ends_at: string;
  status: string;
  total: number;
  available: number;
  sold: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: number;
  order_no: string;
  user_id: number;
  activity_id: number;
  quantity: number;
  unit_price_cents: number;
  total_price_cents: number;
  status: string;
  created_at: string;
  updated_at: string;
  cancelled_at?: string;
}

export interface Page<T> {
  items: T[];
  limit: number;
  offset: number;
}

export interface HealthStatus {
  status: string;
  dependencies?: { postgresql: string };
}
