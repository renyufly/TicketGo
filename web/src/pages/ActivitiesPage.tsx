import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { listActivities } from "../features/activities/api";
import { formatMoney, formatTime } from "../api/format";
import { EmptyState, ErrorAlert, LoadingPanel } from "../components/Feedback";
import { PageHeader, StatusBadge } from "../components/Entity";
import { Pagination } from "../components/Pagination";
const LIMIT = 12;
export function ActivitiesPage() {
  const { token, user } = useAuth();
  const [offset, setOffset] = useState(0);
  const query = useQuery({
    queryKey: ["activities", offset],
    queryFn: () => listActivities(token!, LIMIT, offset),
    enabled: Boolean(token),
    refetchInterval: 15000,
  });
  return (
    <>
      <PageHeader
        eyebrow="实时库存"
        title="秒杀活动"
        actions={
          user?.role === "admin" && (
            <Link className="button" to="/admin/activities/new">
              创建活动
            </Link>
          )
        }
      />
      {query.isPending && <LoadingPanel />}
      {query.error && <ErrorAlert error={query.error} />}
      {query.data?.items.length === 0 && (
        <EmptyState>当前页暂无活动。</EmptyState>
      )}
      <div className="card-grid">
        {query.data?.items.map((activity) => (
          <article className="card activity-card" key={activity.id}>
            <div className="card-top">
              <StatusBadge value={activity.status} />
              <span>
                剩余 {activity.available}/{activity.total}
              </span>
            </div>
            <h2>
              <Link to={`/activities/${activity.id}`}>{activity.name}</Link>
            </h2>
            <strong className="price">
              {formatMoney(activity.price_cents)}
            </strong>
            <div className="stock-meter">
              <span
                style={{
                  width: `${activity.total ? (activity.available / activity.total) * 100 : 0}%`,
                }}
              />
            </div>
            <small>{formatTime(activity.starts_at)} 开始</small>
          </article>
        ))}
      </div>
      {query.data && (
        <Pagination
          offset={offset}
          limit={LIMIT}
          count={query.data.items.length}
          onChange={setOffset}
        />
      )}
    </>
  );
}
