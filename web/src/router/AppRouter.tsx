import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "../components/Layout";
import { ProtectedRoute } from "./ProtectedRoute";
import { LoginPage } from "../pages/LoginPage";
import { RegisterPage } from "../pages/RegisterPage";
import { ProfilePage } from "../pages/ProfilePage";
import { ItemsPage } from "../pages/ItemsPage";
import { ItemDetailPage } from "../pages/ItemDetailPage";
import { CreateItemPage } from "../pages/CreateItemPage";
import { ActivitiesPage } from "../pages/ActivitiesPage";
import { ActivityDetailPage } from "../pages/ActivityDetailPage";
import { CreateActivityPage } from "../pages/CreateActivityPage";
import { OrdersPage } from "../pages/OrdersPage";
import { OrderDetailPage } from "../pages/OrderDetailPage";
import { SystemStatusPage } from "../pages/SystemStatusPage";
import { ForbiddenPage } from "../pages/ForbiddenPage";
import { NotFoundPage } from "../pages/NotFoundPage";

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route index element={<Navigate to="/activities" replace />} />
          <Route path="/me" element={<ProfilePage />} />
          <Route path="/items" element={<ItemsPage />} />
          <Route path="/items/:id" element={<ItemDetailPage />} />
          <Route path="/activities" element={<ActivitiesPage />} />
          <Route path="/activities/:id" element={<ActivityDetailPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route path="/orders/:id" element={<OrderDetailPage />} />
          <Route path="/status" element={<SystemStatusPage />} />
          <Route path="/forbidden" element={<ForbiddenPage />} />
          <Route element={<ProtectedRoute adminOnly />}>
            <Route path="/admin/items/new" element={<CreateItemPage />} />
            <Route
              path="/admin/activities/new"
              element={<CreateActivityPage />}
            />
          </Route>
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
