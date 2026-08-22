/* global process, fetch, console */

const baseUrl = process.env.TICKETGO_E2E_BASE_URL ?? "http://localhost:8080";
const suffix = `${Date.now()}-${Math.random().toString(16).slice(2)}`;

async function request(path, { token, ...options } = {}) {
  const response = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers: {
      Accept: "application/json",
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(
      `${options.method ?? "GET"} ${path}: ${response.status} ${payload.code} request_id=${payload.request_id}`,
    );
  }
  return payload.data;
}

async function registerAndLogin(email, role) {
  await request("/api/v1/users", {
    method: "POST",
    body: JSON.stringify({ email, password: "phase1b-demo-password", role }),
  });
  const login = await request("/api/v1/login", {
    method: "POST",
    body: JSON.stringify({ email, password: "phase1b-demo-password" }),
  });
  return login.access_token;
}

await request("/health/live");
await request("/health/ready");
const adminToken = await registerAndLogin(
  `phase1b-admin-${suffix}@example.test`,
  "admin",
);
const item = await request("/api/v1/items", {
  token: adminToken,
  method: "POST",
  body: JSON.stringify({
    name: `Phase 1B ticket ${suffix}`,
    description: "real API demo flow",
    price_cents: 12000,
  }),
});
const now = Date.now();
const activity = await request("/api/v1/activities", {
  token: adminToken,
  method: "POST",
  body: JSON.stringify({
    item_id: item.id,
    name: `Phase 1B sale ${suffix}`,
    price_cents: 9900,
    starts_at: new Date(now - 60000).toISOString(),
    ends_at: new Date(now + 3600000).toISOString(),
    status: "active",
    total: 2,
  }),
});
const customerToken = await registerAndLogin(
  `phase1b-customer-${suffix}@example.test`,
  "customer",
);
const order = await request(`/api/v1/activities/${activity.id}/seckill`, {
  token: customerToken,
  method: "POST",
  body: JSON.stringify({ quantity: 1 }),
});
await request(`/api/v1/orders/${order.id}`, { token: customerToken });
await request(`/api/v1/orders/${order.id}/cancel`, {
  token: customerToken,
  method: "POST",
});
const refreshed = await request(`/api/v1/activities/${activity.id}`, {
  token: customerToken,
});
if (refreshed.available !== refreshed.total || refreshed.sold !== 0)
  throw new Error("inventory was not restored after cancellation");
console.log(
  `Phase 1B real API demo passed: item=${item.id} activity=${activity.id} order=${order.id}`,
);
