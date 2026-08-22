import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useParams } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../auth/AuthContext";
import { formatMoney, formatTime, localTimeZone } from "../api/format";
import { getActivity, seckill } from "../features/activities/api";
import {
  DefinitionList,
  Detail,
  PageHeader,
  StatusBadge,
} from "../components/Entity";
import { ErrorAlert, LoadingPanel, SuccessAlert } from "../components/Feedback";
const schema = z.object({ quantity: z.number().int().min(1).max(10) });
type FormData = z.infer<typeof schema>;
export function ActivityDetailPage() {
  const { id = "" } = useParams();
  const { token } = useAuth();
  const client = useQueryClient();
  const [now] = useState(() => Date.now());
  const [orderId, setOrderId] = useState<number>();
  const query = useQuery({
    queryKey: ["activity", id],
    queryFn: () => getActivity(token!, id),
    enabled: Boolean(token && id),
    refetchInterval: 10000,
  });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { quantity: 1 },
  });
  const mutation = useMutation({
    mutationFn: (quantity: number) => seckill(token!, id, quantity),
    onSuccess: async (order) => {
      setOrderId(order.id);
      await Promise.all([
        client.invalidateQueries({ queryKey: ["activity", id] }),
        client.invalidateQueries({ queryKey: ["activities"] }),
        client.invalidateQueries({ queryKey: ["orders"] }),
      ]);
    },
  });
  if (query.isPending) return <LoadingPanel />;
  if (query.error) return <ErrorAlert error={query.error} />;
  const activity = query.data!;
  const unavailable =
    activity.status !== "active" ||
    activity.available <= 0 ||
    now < new Date(activity.starts_at).getTime() ||
    now >= new Date(activity.ends_at).getTime();
  return (
    <>
      <PageHeader eyebrow={`活动 #${activity.id}`} title={activity.name} />
      <div className="two-column">
        <section className="panel">
          <DefinitionList>
            <Detail label="商品 ID">
              <Link to={`/items/${activity.item_id}`}>#{activity.item_id}</Link>
            </Detail>
            <Detail label="活动价">{formatMoney(activity.price_cents)}</Detail>
            <Detail label="状态">
              <StatusBadge value={activity.status} />
            </Detail>
            <Detail label="开始">{formatTime(activity.starts_at)}</Detail>
            <Detail label="结束">{formatTime(activity.ends_at)}</Detail>
            <Detail label="本地时区">{localTimeZone}</Detail>
            <Detail label="库存">
              总计 {activity.total} / 可用 {activity.available} / 已售{" "}
              {activity.sold}
            </Detail>
            <Detail label="库存版本">{activity.version}</Detail>
          </DefinitionList>
        </section>
        <section className="panel purchase-panel">
          <div className="eyebrow">同步秒杀</div>
          <h2>{formatMoney(activity.price_cents)}</h2>
          {orderId && (
            <SuccessAlert>
              秒杀成功。
              <Link to={`/orders/${orderId}`}>查看订单 #{orderId}</Link>
            </SuccessAlert>
          )}
          {mutation.error && <ErrorAlert error={mutation.error} />}
          <form
            onSubmit={handleSubmit((values) =>
              mutation.mutate(values.quantity),
            )}
          >
            <label>
              数量（1–10）
              <input
                type="number"
                min="1"
                max="10"
                {...register("quantity", { valueAsNumber: true })}
              />
            </label>
            {errors.quantity && (
              <span className="field-error">请输入 1–10 的整数</span>
            )}
            <button
              className="button accent"
              disabled={mutation.isPending || unavailable}
            >
              {mutation.isPending
                ? "提交中…"
                : unavailable
                  ? "当前不可秒杀"
                  : "立即秒杀"}
            </button>
          </form>
          <p className="muted">
            按钮状态只改善体验；一人一单和库存正确性由 Go/PostgreSQL 最终校验。
          </p>
        </section>
      </div>
    </>
  );
}
