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

const TOKEN_KEY = "ticketgo.access_token";
export const RECENT_EMAIL_KEY = "ticketgo.recent_email";

interface AuthValue {
  token: string | null;
  user: User | null;
  restoring: boolean;
  login: (email: string, password: string) => Promise<User>;
  logout: () => void;
}

const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() =>
    sessionStorage.getItem(TOKEN_KEY),
  );
  const [user, setUser] = useState<User | null>(null);
  const [restoring, setRestoring] = useState(Boolean(token));

  const logout = useCallback(() => {
    sessionStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
    setRestoring(false);
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(logout);
    return () => setUnauthorizedHandler(undefined);
  }, [logout]);

  useEffect(() => {
    if (!token) return;
    let active = true;
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
    const { data } = await loginUser({ email, password });
    sessionStorage.setItem(TOKEN_KEY, data.access_token);
    localStorage.setItem(RECENT_EMAIL_KEY, email);
    const current = await getCurrentUser(data.access_token);
    setToken(data.access_token);
    setUser(current.data);
    return current.data;
  }, []);

  const value = useMemo(
    () => ({ token, user, restoring, login, logout }),
    [token, user, restoring, login, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Fast refresh warning is intentionally scoped: this hook consumes the context defined above.
// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
