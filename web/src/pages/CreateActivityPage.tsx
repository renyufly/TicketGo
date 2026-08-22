import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../auth/AuthContext";
import { createActivity } from "../features/activities/api";
import { listItems } from "../features/items/api";
import { ErrorAlert } from "../components/Feedback";
import { PageHeader } from "../components/Entity";
import { localTimeZone } from "../api/format";
const schema = z
  .object({
    item_id: z.number().int().min(1),
    name: z.string().min(1).max(200),
    price_cents: z.number().int().min(1).max(1_000_000_000_000),
    starts_at: z.string().min(1),
    ends_at: z.string().min(1),
    status: z.enum(["draft", "active"]),
    total: z.number().int().min(1),
  })
  .refine((v) => new Date(v.ends_at) > new Date(v.starts_at), {
    message: "结束时间必须晚于开始时间",
    path: ["ends_at"],
  });
type FormData = z.infer<typeof schema>;
export function CreateActivityPage() {
  const { token } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<unknown>();
  const items = useQuery({
    queryKey: ["items", "activity-form"],
    queryFn: () => listItems(token!, 100, 0),
  });
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { status: "active", total: 1 },
  });
  const submit = handleSubmit(async (values) => {
    setError(undefined);
    try {
      const activity = await createActivity(token!, {
        ...values,
        starts_at: new Date(values.starts_at).toISOString(),
        ends_at: new Date(values.ends_at).toISOString(),
      });
      navigate(`/activities/${activity.id}`);
    } catch (caught) {
      setError(caught);
    }
  });
  return (
    <>
      <PageHeader eyebrow="管理员" title="创建活动与库存" />
      {error !== undefined && <ErrorAlert error={error} />}
      {items.error && <ErrorAlert error={items.error} />}
      <form className="panel form-grid" onSubmit={submit}>
        <label>
          商品
          <select {...register("item_id", { valueAsNumber: true })}>
            <option value="">请选择商品</option>
            {items.data?.items.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name} (#{item.id})
              </option>
            ))}
          </select>
        </label>
        <label>
          活动名称
          <input {...register("name")} />
        </label>
        <label>
          活动价（cents）
          <input
            type="number"
            min="1"
            {...register("price_cents", { valueAsNumber: true })}
          />
        </label>
        <label>
          总库存
          <input
            type="number"
            min="1"
            {...register("total", { valueAsNumber: true })}
          />
        </label>
        <label>
          开始时间（{localTimeZone}）
          <input type="datetime-local" {...register("starts_at")} />
        </label>
        <label>
          结束时间（{localTimeZone}）
          <input type="datetime-local" {...register("ends_at")} />
        </label>
        {errors.ends_at && (
          <span className="field-error full">{errors.ends_at.message}</span>
        )}
        <label>
          状态
          <select {...register("status")}>
            <option value="active">active</option>
            <option value="draft">draft</option>
          </select>
        </label>
        <div className="full">
          <button
            className="button"
            disabled={isSubmitting || !items.data?.items.length}
          >
            {isSubmitting ? "创建中…" : "创建活动"}
          </button>
        </div>
      </form>
    </>
  );
}
