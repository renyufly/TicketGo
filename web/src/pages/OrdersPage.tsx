// 实现了一个 React 的 “我的订单”列表页面：
// 获取当前用户的订单，并以表格形式展示，
// 同时支持分页、加载状态和错误提示

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { listOrders } from "../features/orders/api";
import { formatMoney, formatTime } from "../api/format";
import { EmptyState, ErrorAlert, LoadingPanel } from "../components/Feedback";
import { PageHeader, StatusBadge } from "../components/Entity";
import { Pagination } from "../components/Pagination";

// LIMIT = 10：每页最多 10 条订单
// token：当前登录用户的 Token
// offset：当前分页偏移量
const LIMIT = 10;
export function OrdersPage() {
  const { token } = useAuth();
  const [offset, setOffset] = useState(0);

  // 请求订单数据: React Query 请求后端
  /* queryKey: ["orders", offset] 表示不同 offset 是不同的查询.
     offset 改变后，React Query 会自动重新执行listOrders()
     ["orders", 0] → 第一页  ["orders", 10] → 第二页
  */
  // enabled: Boolean(token) 表示只有存在 Token 时才发送请求
  const query = useQuery({
    queryKey: ["orders", offset],
    queryFn: () => listOrders(token!, LIMIT, offset),
    enabled: Boolean(token),
  });
  return (
    <>
      <PageHeader eyebrow="仅当前账号可见" title="我的订单" />
      {/* 根据请求状态显示不同 UI:加载中、请求失败、无订单 */}
      {query.isPending && <LoadingPanel />}
      {query.error && <ErrorAlert error={query.error} />}
      {query.data?.items.length === 0 && (
        <EmptyState>还没有订单，请先参加秒杀活动。</EmptyState>
      )}
      {/* 有订单就显示表格 */}
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
              {/* 遍历订单数组，每个 order 生成一行 */}
              {query.data.items.map((order) => (
                <tr key={order.id}>
                  <td>
                    {/* 订单号还可以点击，进入订单详情页 */}
                    <Link to={`/orders/${order.id}`}>{order.order_no}</Link>
                  </td>
                  <td>
                    <Link to={`/activities/${order.activity_id}`}>
                      #{order.activity_id}
                    </Link>
                  </td>
                  <td>{order.quantity}</td>
                  {/* 格式化数据显示 */}
                  <td>{formatMoney(order.total_price_cents)}</td>
                  <td>
                    {/* 把 pending / cancelled / paid 等状态显示成统一的状态标签 */}
                    <StatusBadge value={order.status} />
                  </td>
                  {/* 负责格式化时间 */}
                  <td>{formatTime(order.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {/* 用户点击下一页后，Pagination 调用setOffset(10)， 
      导致 offset 改变，React Query 重新请求，页面重新渲染 */}
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
