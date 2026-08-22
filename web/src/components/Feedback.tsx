import type { ReactNode } from "react";
import { ApiError } from "../api/client";

export function LoadingPanel({ text = "加载中…" }: { text?: string }) {
  return (
    <div className="panel state-panel" aria-live="polite">
      {text}
    </div>
  );
}

export function EmptyState({ children }: { children: ReactNode }) {
  return <div className="panel state-panel muted">{children}</div>;
}

export function ErrorAlert({ error }: { error: unknown }) {
  const apiError = error instanceof ApiError ? error : undefined;
  return (
    <div className="alert alert-error" role="alert">
      <strong>
        {apiError ? userMessage(apiError) : "操作失败，请稍后重试。"}
      </strong>
      {apiError?.requestId && (
        <div className="request-id">
          追踪编号：<code>{apiError.requestId}</code>
          <button
            type="button"
            className="link-button"
            onClick={() =>
              void navigator.clipboard?.writeText(apiError.requestId!)
            }
          >
            复制
          </button>
        </div>
      )}
    </div>
  );
}

function userMessage(error: ApiError) {
  const messages: Record<string, string> = {
    unauthenticated: "登录状态已失效，请重新登录。",
    forbidden: "当前账号没有执行此操作的权限。",
    user_disabled: "当前账号已被停用。",
    out_of_stock: "库存不足，本次秒杀未成功。",
    activity_unavailable: "活动尚未开始、已经结束或当前不可用。",
    conflict: "请求与当前状态冲突，可能已参加过活动或资源已发生变化。",
    dependency_unavailable: "PostgreSQL 暂不可用，请稍后重试。",
    validation_error: "输入内容未通过校验，请检查后重试。",
    invalid_request: "请求内容不正确，请检查后重试。",
    not_found: "没有找到对应数据。",
    network_error: error.message,
  };
  return messages[error.code] ?? error.message;
}

export function SuccessAlert({ children }: { children: ReactNode }) {
  return (
    <div className="alert alert-success" role="status">
      {children}
    </div>
  );
}
