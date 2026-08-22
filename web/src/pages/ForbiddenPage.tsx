import { Link } from "react-router-dom";
export function ForbiddenPage() {
  return (
    <section className="panel state-panel">
      <div className="eyebrow">403</div>
      <h1>权限不足</h1>
      <p>当前账号不是管理员，后端也会拒绝管理接口。</p>
      <Link className="button" to="/activities">
        返回活动列表
      </Link>
    </section>
  );
}
