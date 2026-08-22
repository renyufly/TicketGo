import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../auth/AuthContext";
import { createItem } from "../features/items/api";
import { ErrorAlert } from "../components/Feedback";
import { PageHeader } from "../components/Entity";
const schema = z.object({
  name: z.string().min(1).max(200),
  description: z.string().max(2000),
  price_cents: z.number().int().min(1).max(1_000_000_000_000),
  status: z.enum(["active", "inactive"]),
});
type FormData = z.infer<typeof schema>;
export function CreateItemPage() {
  const { token } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<unknown>();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { description: "", status: "active" },
  });
  const submit = handleSubmit(async (values) => {
    setError(undefined);
    try {
      const item = await createItem(token!, values);
      navigate(`/items/${item.id}`);
    } catch (caught) {
      setError(caught);
    }
  });
  return (
    <>
      <PageHeader eyebrow="管理员" title="创建商品" />
      {error !== undefined && <ErrorAlert error={error} />}
      <form className="panel form-grid" onSubmit={submit}>
        <label>
          名称
          <input {...register("name")} />
        </label>
        {errors.name && (
          <span className="field-error">请输入 1–200 字名称</span>
        )}
        <label className="full">
          描述
          <textarea rows={4} {...register("description")} />
        </label>
        <label>
          价格（cents）
          <input
            type="number"
            min="1"
            {...register("price_cents", { valueAsNumber: true })}
          />
        </label>
        <label>
          状态
          <select {...register("status")}>
            <option value="active">active</option>
            <option value="inactive">inactive</option>
          </select>
        </label>
        <div className="full">
          <button className="button" disabled={isSubmitting}>
            {isSubmitting ? "创建中…" : "创建商品"}
          </button>
        </div>
      </form>
    </>
  );
}
