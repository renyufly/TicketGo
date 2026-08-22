interface SuccessEnvelope<T> {
  data: T;
  request_id: string;
}

interface ErrorEnvelope {
  code?: string;
  message?: string;
  request_id?: string;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

let onUnauthorized: (() => void) | undefined;

export function setUnauthorizedHandler(handler?: () => void) {
  onUnauthorized = handler;
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<{ data: T; requestId: string }> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body) headers.set("Content-Type", "application/json");
  if (options.token) headers.set("Authorization", `Bearer ${options.token}`);

  let response: Response;
  try {
    response = await fetch(path, { ...options, headers });
  } catch {
    throw new ApiError(
      0,
      "network_error",
      "无法连接后端服务，请确认 Gin 已启动。",
    );
  }

  const payload = (await response
    .json()
    .catch(() => ({}))) as SuccessEnvelope<T> & ErrorEnvelope;
  if (!response.ok) {
    const error = new ApiError(
      response.status,
      payload.code ?? "unknown_error",
      payload.message ?? "请求失败，请稍后重试。",
      payload.request_id,
    );
    if (response.status === 401) onUnauthorized?.();
    throw error;
  }
  return { data: payload.data, requestId: payload.request_id };
}
