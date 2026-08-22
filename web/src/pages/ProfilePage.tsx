import {
  DefinitionList,
  Detail,
  PageHeader,
  StatusBadge,
} from "../components/Entity";
import { useAuth } from "../auth/AuthContext";
import { formatTime } from "../api/format";

export function ProfilePage() {
  const { user } = useAuth();
  if (!user) return null;
  return (
    <>
      <PageHeader eyebrow="认证信息" title="当前用户" />
      <section className="panel">
        <DefinitionList>
          <Detail label="用户 ID">{user.id}</Detail>
          <Detail label="邮箱">{user.email}</Detail>
          <Detail label="角色">
            <StatusBadge value={user.role} />
          </Detail>
          <Detail label="账号状态">
            <StatusBadge value={user.status} />
          </Detail>
          <Detail label="创建时间">{formatTime(user.created_at)}</Detail>
        </DefinitionList>
      </section>
    </>
  );
}
