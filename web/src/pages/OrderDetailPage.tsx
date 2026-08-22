import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { formatMoney, formatTime } from "../api/format";
import { cancelOrder, getOrder } from "../features/orders/api";
import {
  DefinitionList,
  Detail,
  PageHeader,
  StatusBadge,
} from "../components/Entity";
import { ErrorAlert, LoadingPanel, SuccessAlert } from "../components/Feedback";
export function OrderDetailPage() {
  const { id = "" } = useParams();
  const { token } = useAuth();
  const client = useQueryClient();
  const query = useQuery({
    queryKey: ["order", id],
    queryFn: () => getOrder(token!, id),
    enabled: Boolean(token && id),
  });
  const mutation = useMutation({
    mutationFn: () => cancelOrder(token!, Number(id)),
    onSuccess: async (order) => {
      client.setQueryData(["order", id], order);
      await Promise.all([
        client.invalidateQueries({ queryKey: ["orders"] }),
        client.invalidateQueries({ queryKey: ["activities"] }),
        client.invalidateQueries({
          queryKey: ["activity", String(order.activity_id)],
        }),
      ]);
    },
  });
  if (query.isPending) return <LoadingPanel />;
  if (query.error) return <ErrorAlert error={query.error} />;
  const order = query.data!;
  return (
    <>
      <PageHeader eyebrow={`订单 #${order.id}`} title={order.order_no} />
      {mutation.isSuccess && (
        <SuccessAlert>订单已取消，活动库存查询已刷新。</SuccessAlert>
      )}
      {mutation.error && <ErrorAlert error={mutation.error} />}
      <section className="panel">
        <DefinitionList>
          <Detail label="状态">
            <StatusBadge value={order.status} />
          </Detail>
          <Detail label="活动">
            <Link to={`/activities/${order.activity_id}`}>
              #{order.activity_id}
            </Link>
          </Detail>
          <Detail label="数量">{order.quantity}</Detail>
          <Detail label="单价">{formatMoney(order.unit_price_cents)}</Detail>
          <Detail label="总额">{formatMoney(order.total_price_cents)}</Detail>
          <Detail label="创建时间">{formatTime(order.created_at)}</Detail>
          <Detail label="更新时间">{formatTime(order.updated_at)}</Detail>
          {order.cancelled_at && (
            <Detail label="取消时间">{formatTime(order.cancelled_at)}</Detail>
          )}
        </DefinitionList>
        {order.status === "pending" && (
          <button
            type="button"
            className="button danger"
            disabled={mutation.isPending}
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending ? "取消中…" : "取消订单并回补库存"}
          </button>
        )}
      </section>
    </>
  );
}
