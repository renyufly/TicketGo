import { Link } from "react-router-dom";
export function NotFoundPage() {
  return (
    <main className="auth-page">
      <section className="auth-card">
        <div className="eyebrow">404</div>
        <h1>页面不存在</h1>
        <Link to="/activities">返回 TicketGo</Link>
      </section>
    </main>
  );
}
