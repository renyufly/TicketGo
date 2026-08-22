// 一个简单的 React 分页组件，
// 核心就是通过修改 offset 来切换上一页/下一页

// offset = 20   当前从第 20 条开始
// limit  = 10   每页 10 条
// count  = 10   当前页实际拿到了 10 条
// onChange 是父组件传进来的修改 offset 的函数
export function Pagination({
  offset,
  limit,
  count,
  onChange,
}: {
  offset: number;
  limit: number;
  count: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="pagination" aria-label="分页">
      {/* 当第一页 offset = 0 时，禁用“上一页”
      点击时，触发setOffset() 返回上一页 */}
      <button
        type="button"
        className="button secondary"
        disabled={offset === 0}
        onClick={() => onChange(Math.max(0, offset - limit))}
      >
        上一页
      </button>
      <span>第 {Math.floor(offset / limit) + 1} 页</span>
      {/* 父组件中的 offset 改变，React Query 就会请求下一页数据.
      下一页按钮在两种情况下禁用 */}
      <button
        type="button"
        className="button secondary"
        disabled={count < limit || offset + limit > 10000}
        onClick={() => onChange(offset + limit)}
      >
        下一页
      </button>
    </div>
  );
}
