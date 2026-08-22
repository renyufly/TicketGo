// 封装了一个统一的前端 HTTP 请求工具 apiRequest，
// 负责发请求、带 Token、解析响应和统一处理错误
// 整个前端访问 Gin 后端的统一 HTTP 客户端封装，
// 让具体的订单、用户等 API 只需要关心“请求哪个接口”，
// 不用重复处理认证和错误

/*
业务代码 -> apiRequest() -> 添加 JSON 请求头
-> 添加 Bearer Token -> fetch() -> 解析 JSON
-> 失败 → ApiError
           └─ 401 → onUnauthorized()
-> 成功 -> 返回 data + requestId    */

/* 定义后端响应格式： T 是泛型，代表 data 可以是任意类型
{
  "data": {...},
  "request_id": "abc123"
}
*/
interface SuccessEnvelope<T> {
  data: T;
  request_id: string;
}

// 表示失败时可能返回错误码、错误信息和请求 ID
interface ErrorEnvelope {
  code?: string;
  message?: string;
  request_id?: string;
}

/* 自己定义一种 API 错误，在普通 Error 基础上额外保存：
status：HTTP 状态码，例如 401、404
code：业务错误码
message：错误描述
requestId：这次请求的 ID，方便日志排查
*/
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

// 允许外部注册一个函数，
// 之后只要接口返回 401 未登录/Token 失效，就自动调用它
let onUnauthorized: (() => void) | undefined;

export function setUnauthorizedHandler(handler?: () => void) {
  onUnauthorized = handler;
}

// 对原生 fetch() 的统一封装
export async function apiRequest<T>(
  path: string,
  options: RequestInit & { token?: string } = {},
): Promise<{ data: T; requestId: string }> {
  // 准备请求头：希望后端返回 JSON
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");

  // 如果有请求体，告诉后端发送的数据是 JSON
  if (options.body) headers.set("Content-Type", "application/json");
  // 如果传入 Token，供 Gin 后端 JWT 中间件验证身份
  if (options.token) headers.set("Authorization", `Bearer ${options.token}`);

  let response: Response;
  try {
    // 真正发送请求
    response = await fetch(path, { ...options, headers });
  } catch {
    throw new ApiError(
      0,
      "network_error",
      "无法连接后端服务，请确认 Gin 已启动。",
    );
  }

  // 解析后端 JSON
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
