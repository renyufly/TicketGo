import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { getItem } from "../features/items/api";
import { formatMoney, formatTime } from "../api/format";
import {
  DefinitionList,
  Detail,
  PageHeader,
  StatusBadge,
} from "../components/Entity";
import { ErrorAlert, LoadingPanel } from "../components/Feedback";

export function ItemDetailPage() {
  const { id = "" } = useParams();
  const { token } = useAuth();
  const query = useQuery({
    queryKey: ["item", id],
    queryFn: () => getItem(token!, id),
    enabled: Boolean(token && id),
  });
  if (query.isPending) return <LoadingPanel />;
  if (query.error) return <ErrorAlert error={query.error} />;
  const item = query.data!;
  return (
    <>
      <PageHeader eyebrow={`商品 #${item.id}`} title={item.name} />
      <section className="panel">
        <p>{item.description || "暂无描述"}</p>
        <DefinitionList>
          <Detail label="价格">{formatMoney(item.price_cents)}</Detail>
          <Detail label="状态">
            <StatusBadge value={item.status} />
          </Detail>
          <Detail label="创建时间">{formatTime(item.created_at)}</Detail>
          <Detail label="更新时间">{formatTime(item.updated_at)}</Detail>
        </DefinitionList>
      </section>
    </>
  );
}
