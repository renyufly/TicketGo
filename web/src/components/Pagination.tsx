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
      <button
        type="button"
        className="button secondary"
        disabled={offset === 0}
        onClick={() => onChange(Math.max(0, offset - limit))}
      >
        上一页
      </button>
      <span>第 {Math.floor(offset / limit) + 1} 页</span>
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
