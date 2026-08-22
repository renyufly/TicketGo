import { ApiError, apiRequest } from "./client";

test("apiRequest unwraps the standard success envelope", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue(
        new Response(
          JSON.stringify({ data: { id: 7 }, request_id: "req-ok" }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      ),
  );
  await expect(apiRequest<{ id: number }>("/api/v1/example")).resolves.toEqual({
    data: { id: 7 },
    requestId: "req-ok",
  });
});

test("apiRequest preserves stable error code and request id", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          code: "out_of_stock",
          message: "stock unavailable",
          request_id: "req-fail",
        }),
        { status: 409, headers: { "Content-Type": "application/json" } },
      ),
    ),
  );
  await expect(apiRequest("/api/v1/example")).rejects.toMatchObject({
    status: 409,
    code: "out_of_stock",
    requestId: "req-fail",
  });
  expect(ApiError).toBeDefined();
});
