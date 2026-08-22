import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";
import { z } from "zod";
import { ErrorAlert, SuccessAlert } from "../components/Feedback";
import { registerUser } from "../features/users/api";
import type { UserRole } from "../api/types";

const adminEnabled = import.meta.env.VITE_ALLOW_ADMIN_REGISTRATION === "true";
const schema = z.object({
  email: z.email("请输入有效邮箱"),
  password: z.string().min(8, "密码至少 8 位").max(72),
  role: z.enum(["customer", "admin"]),
});
type FormData = z.infer<typeof schema>;

export function RegisterPage() {
  const [error, setError] = useState<unknown>();
  const [created, setCreated] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { role: "customer" },
  });
  const onSubmit = handleSubmit(async (values) => {
    setError(undefined);
    setCreated(false);
    try {
      await registerUser({ ...values, role: values.role as UserRole });
      setCreated(true);
    } catch (caught) {
      setError(caught);
    }
  });
  return (
    <main className="auth-page">
      <section className="auth-card">
        <div className="eyebrow">创建演示身份</div>
        <h1>注册账号</h1>
        {created && <SuccessAlert>注册成功，请返回登录。</SuccessAlert>}
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
              autoComplete="new-password"
              {...register("password")}
            />
          </label>
          {errors.password && (
            <span className="field-error">
              {String(errors.password.message)}
            </span>
          )}
          <fieldset>
            <legend>角色</legend>
            <label className="radio">
              <input type="radio" value="customer" {...register("role")} />{" "}
              普通用户 customer
            </label>
            {adminEnabled && (
              <label className="radio">
                <input type="radio" value="admin" {...register("role")} />{" "}
                管理员 admin（仅本地演示）
              </label>
            )}
          </fieldset>
          {adminEnabled && (
            <p className="warning">
              匿名管理员注册是本地演示开关，生产环境禁止开启。
            </p>
          )}
          <button className="button" disabled={isSubmitting}>
            {isSubmitting ? "注册中…" : "注册"}
          </button>
        </form>
        <p>
          <Link to="/login">返回登录</Link>
        </p>
      </section>
    </main>
  );
}
