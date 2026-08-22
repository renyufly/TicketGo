import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { listItems } from "../features/items/api";
import { formatMoney } from "../api/format";
import { EmptyState, ErrorAlert, LoadingPanel } from "../components/Feedback";
import { PageHeader, StatusBadge } from "../components/Entity";
import { Pagination } from "../components/Pagination";

const LIMIT = 12;
export function ItemsPage() {
  const { token, user } = useAuth();
  const [offset, setOffset] = useState(0);
  const query = useQuery({
    queryKey: ["items", offset],
    queryFn: () => listItems(token!, LIMIT, offset),
    enabled: Boolean(token),
  });
  return (
    <>
      <PageHeader
        eyebrow="资源目录"
        title="商品 / 门票"
        actions={
          user?.role === "admin" && (
            <Link className="button" to="/admin/items/new">
              创建商品
            </Link>
          )
        }
      />
      {query.isPending && <LoadingPanel />}
      {query.error && <ErrorAlert error={query.error} />}
      {query.data && query.data.items.length === 0 && (
        <EmptyState>当前页暂无商品。</EmptyState>
      )}
      {query.data && query.data.items.length > 0 && (
        <div className="card-grid">
          {query.data.items.map((item) => (
            <article className="card" key={item.id}>
              <div className="card-top">
                <StatusBadge value={item.status} />
                <span>#{item.id}</span>
              </div>
              <h2>
                <Link to={`/items/${item.id}`}>{item.name}</Link>
              </h2>
              <p>{item.description || "暂无描述"}</p>
              <strong className="price">{formatMoney(item.price_cents)}</strong>
            </article>
          ))}
        </div>
      )}
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
