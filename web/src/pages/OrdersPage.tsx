import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { listOrders } from "../features/orders/api";
import { formatMoney, formatTime } from "../api/format";
import { EmptyState, ErrorAlert, LoadingPanel } from "../components/Feedback";
import { PageHeader, StatusBadge } from "../components/Entity";
import { Pagination } from "../components/Pagination";
const LIMIT = 10;
export function OrdersPage() {
  const { token } = useAuth();
  const [offset, setOffset] = useState(0);
  const query = useQuery({
    queryKey: ["orders", offset],
    queryFn: () => listOrders(token!, LIMIT, offset),
    enabled: Boolean(token),
  });
  return (
    <>
      <PageHeader eyebrow="仅当前账号可见" title="我的订单" />
      {query.isPending && <LoadingPanel />}
      {query.error && <ErrorAlert error={query.error} />}
      {query.data?.items.length === 0 && (
        <EmptyState>还没有订单，请先参加秒杀活动。</EmptyState>
      )}
      {query.data && query.data.items.length > 0 && (
        <div className="table-wrap panel">
          <table>
            <thead>
              <tr>
                <th>订单号</th>
                <th>活动</th>
                <th>数量</th>
                <th>金额</th>
                <th>状态</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              {query.data.items.map((order) => (
                <tr key={order.id}>
                  <td>
                    <Link to={`/orders/${order.id}`}>{order.order_no}</Link>
                  </td>
                  <td>
                    <Link to={`/activities/${order.activity_id}`}>
                      #{order.activity_id}
                    </Link>
                  </td>
                  <td>{order.quantity}</td>
                  <td>{formatMoney(order.total_price_cents)}</td>
                  <td>
                    <StatusBadge value={order.status} />
                  </td>
                  <td>{formatTime(order.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
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
