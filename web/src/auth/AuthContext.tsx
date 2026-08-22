// React 项目的全局登录认证管理 AuthContext，
// 负责保存 Token、当前用户，以及登录、登出、自动恢复登录状态

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { getCurrentUser, loginUser } from "../features/users/api";
import { setUnauthorizedHandler } from "../api/client";
import type { User } from "../api/types";

// 定义浏览器存储时使用的 key（键名）
// 注意：没有 export 的只能在当前文件内部使用
const TOKEN_KEY = "ticketgo.access_token";
export const RECENT_EMAIL_KEY = "ticketgo.recent_email";

/* 定义认证状态：
token：JWT Token
user：当前登录用户
restoring：是否正在恢复登录状态
login()：登录
logout()：退出登录
*/
interface AuthValue {
  token: string | null;
  user: User | null;
  restoring: boolean;
  login: (email: string, password: string) => Promise<User>;
  logout: () => void;
}

// 创建了一个全局认证数据容器
const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  // 初始化 Token：页面刷新后，
  // 从浏览器 sessionStorage 读取之前保存的 Token。
  // 所以刷新网页不会立刻丢失登录状态
  const [token, setToken] = useState<string | null>(() =>
    sessionStorage.getItem(TOKEN_KEY),
  );
  const [user, setUser] = useState<User | null>(null);
  const [restoring, setRestoring] = useState(Boolean(token));

  /* 退出登录:
  删除 Token -> token = null -> user = null -> 变成未登录状态
  useCallback 主要是让 logout 函数引用保持稳定，避免不必要的重新创建
  */
  const logout = useCallback(() => {
    sessionStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
    setRestoring(false);
  }, []);

  // 如果后端返回：401 Unauthorized
  // Token 过期 → 后端返回 401 → 前端自动退出登录
  useEffect(() => {
    setUnauthorizedHandler(logout);
    return () => setUnauthorizedHandler(undefined);
  }, [logout]);

  useEffect(() => {
    if (!token) return;
    // active 防止组件已经卸载后，异步请求回来还继续 setUser()
    let active = true;
    // 如果浏览器里已经有 Token：向后端请求：这个 Token 对应哪个用户？
    // 成功则 恢复用户信息. 否则Token 可能已经失效，直接退出登录
    getCurrentUser(token)
      .then(({ data }) => {
        if (active) setUser(data);
      })
      .catch(() => {
        if (active) logout();
      })
      .finally(() => {
        if (active) setRestoring(false);
      });
    return () => {
      active = false;
    };
  }, [token, logout]);

  const login = useCallback(async (email: string, password: string) => {
    // 后端验证账号密码
    const { data } = await loginUser({ email, password });
    // 获得 access_token，保存到 sessionStorage
    sessionStorage.setItem(TOKEN_KEY, data.access_token);
    // 额外记住最近登录的邮箱，通常用于下次自动填充登录框
    localStorage.setItem(RECENT_EMAIL_KEY, email);
    // 获得当前用户信息
    const current = await getCurrentUser(data.access_token);
    setToken(data.access_token);
    setUser(current.data);
    return current.data;
  }, []);

  /*
  把这些认证数据组合成一个对象，并避免每次渲染都无意义地创建新对象.
  然后提供给整个子组件树.
  */
  const value = useMemo(
    () => ({ token, user, restoring, login, logout }),
    [token, user, restoring, login, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

/* useAuth 是 自定义Hook，不是组件：
const { token, user, login, logout } = useAuth();
就是对下面代码的封装.
useAuth() 是对 useContext(AuthContext) 的一层安全封装，
让其他组件能方便地获取 token、user、login、logout，
并在没有 AuthProvider 时立即报错
*/
// Fast refresh warning is intentionally scoped: this hook consumes the context defined above.
// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  // 从最近的 AuthContext.Provider 中读取认证数据：
  /*{
    token,
    user,
    restoring,
    login,
    logout
    }  */
  const context = useContext(AuthContext);
  // 检查当前组件有没有被 AuthProvider 包住
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
/* 注意：
如果没有 useAuth，每个组件都得：
const auth = useContext(AuthContext);
if (!auth) {
  throw new Error(...);
}

有了useAuth以后只需要：
const { user, token } = useAuth();
更简洁，而且“必须在 AuthProvider 内使用”的检查只需要写一次
*/
