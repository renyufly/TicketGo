import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

export function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const signOut = () => {
    logout();
    navigate("/login");
  };
  return (
    <div className="app-shell">
      <header className="topbar">
        <NavLink to="/activities" className="brand">
          TicketGo <span>演示台</span>
        </NavLink>
        <nav aria-label="主导航">
          <NavLink to="/items">商品</NavLink>
          <NavLink to="/activities">活动</NavLink>
          <NavLink to="/orders">订单</NavLink>
          <NavLink to="/status">系统状态</NavLink>
          {user?.role === "admin" && (
            <NavLink to="/admin/items/new">管理</NavLink>
          )}
        </nav>
        <div className="account">
          <NavLink to="/me">{user?.email}</NavLink>
          <span className="badge">{user?.role}</span>
          <button type="button" className="link-button" onClick={signOut}>
            退出 / 切换账号
          </button>
        </div>
      </header>
      <main className="page">
        <Outlet />
      </main>
      <footer>页面状态仅用于演示；库存、权限与订单结果始终以后端为准。</footer>
    </div>
  );
}
