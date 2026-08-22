// React 项目的入口文件
// 初始化一些全局功能，然后启动整个 React 应用

/*
StrictMode：React 开发模式检查，帮助发现潜在问题。
createRoot：把 React 应用挂载到 HTML 页面。
BrowserRouter：提供前端路由功能，比如 /login、/orders。
QueryClient：React Query 的核心对象，负责请求缓存、重试等。
QueryClientProvider：让整个项目都能使用 React Query。
AuthProvider：提供全局登录/用户认证状态。
AppRouter：项目自己的路由配置。
global.css：全局样式。
*/
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "./auth/AuthContext";
import { AppRouter } from "./router/AppRouter";
import "./styles/global.css";

// 设置所有 API 请求的默认行为：
// staleTime: 5000 → 数据 5 秒内认为是新鲜的
// retry: 1       → 请求失败自动重试 1 次
// mutations-retry: false  → POST/PUT/DELETE 等失败后不自动重试
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 5000, retry: 1 },
    mutations: { retry: false },
  },
});

// 启动React，找到 HTML 中 <div id="root"></div>
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <AppRouter />
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
);
