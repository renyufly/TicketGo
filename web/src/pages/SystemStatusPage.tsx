import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "../api/client";
import type { HealthStatus } from "../api/types";
import { ErrorAlert, LoadingPanel } from "../components/Feedback";
import { PageHeader, StatusBadge } from "../components/Entity";

function HealthCard({
  title,
  path,
  description,
}: {
  title: string;
  path: string;
  description: string;
}) {
  const query = useQuery({
    queryKey: ["health", path],
    queryFn: async () => (await apiRequest<HealthStatus>(path)).data,
    refetchInterval: 5000,
    retry: false,
  });
  return (
    <article className="panel health-card">
      <div>
        <div className="eyebrow">{path}</div>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
      {query.isPending ? (
        <LoadingPanel text="检查中…" />
      ) : query.error ? (
        <ErrorAlert error={query.error} />
      ) : (
        <div>
          <StatusBadge value={query.data.status} />
          {query.data.dependencies?.postgresql && (
            <p>PostgreSQL：{query.data.dependencies.postgresql}</p>
          )}
        </div>
      )}
    </article>
  );
}
export function SystemStatusPage() {
  return (
    <>
      <PageHeader eyebrow="每 5 秒自动刷新" title="系统状态" />
      <div className="two-column">
        <HealthCard
          title="进程存活"
          path="/health/live"
          description="仅确认 Gin 进程可以响应，不访问外部依赖。"
        />
        <HealthCard
          title="服务就绪"
          path="/health/ready"
          description="确认 PostgreSQL 可用；依赖故障时返回 503。"
        />
      </div>
    </>
  );
}
