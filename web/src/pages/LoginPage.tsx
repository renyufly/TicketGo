import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { z } from "zod";
import { RECENT_EMAIL_KEY, useAuth } from "../auth/AuthContext";
import { ErrorAlert } from "../components/Feedback";

const schema = z.object({
  email: z.email("请输入有效邮箱"),
  password: z.string().min(8, "密码至少 8 位").max(72),
});
type FormData = z.infer<typeof schema>;

export function LoginPage() {
  const { token, user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState<unknown>();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: localStorage.getItem(RECENT_EMAIL_KEY) ?? "",
      password: "",
    },
  });
  if (token && user) return <Navigate to="/activities" replace />;
  const onSubmit = handleSubmit(async (values) => {
    setError(undefined);
    try {
      const current = await login(values.email, values.password);
      const from = (location.state as { from?: string } | null)?.from;
      navigate(
        from ?? (current.role === "admin" ? "/admin/items/new" : "/activities"),
        { replace: true },
      );
    } catch (caught) {
      setError(caught);
    }
  });
  return (
    <main className="auth-page">
      <section className="auth-card">
        <div className="eyebrow">高并发秒杀系统</div>
        <h1>登录 TicketGo</h1>
        <p>JWT 仅保存在当前标签页的 sessionStorage，不保存密码。</p>
        {error !== undefined && <ErrorAlert error={error} />}
        <form onSubmit={onSubmit}>
          <label>
            邮箱
            <input type="email" autoComplete="email" {...register("email")} />
          </label>
          {errors.email && (
            <span className="field-error">{String(errors.email.message)}</span>
          )}
          <label>
            密码
            <input
              type="password"
              autoComplete="current-password"
              {...register("password")}
            />
          </label>
          {errors.password && (
            <span className="field-error">
              {String(errors.password.message)}
            </span>
          )}
          <button className="button" disabled={isSubmitting}>
            {isSubmitting ? "登录中…" : "登录"}
          </button>
        </form>
        <p>
          还没有账号？ <Link to="/register">注册演示账号</Link>
        </p>
      </section>
    </main>
  );
}
