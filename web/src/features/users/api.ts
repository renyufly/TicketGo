import { apiRequest } from "../../api/client";
import type { User, UserRole } from "../../api/types";

export const registerUser = (input: {
  email: string;
  password: string;
  role: UserRole;
}) =>
  apiRequest<User>("/api/v1/users", {
    method: "POST",
    body: JSON.stringify(input),
  });

export const loginUser = (input: { email: string; password: string }) =>
  apiRequest<{ access_token: string; token_type: string }>("/api/v1/login", {
    method: "POST",
    body: JSON.stringify(input),
  });

export const getCurrentUser = (token: string) =>
  apiRequest<User>("/api/v1/users/me", { token });
