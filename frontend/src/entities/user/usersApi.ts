import { request } from "../../shared/http";
import type { User, UserInput } from "./user";

interface UserListResponse {
  items: User[];
}

export const usersApi = {
  async list(): Promise<User[]> {
    const response = await request<UserListResponse>("/api/users");
    if (!Array.isArray(response.items)) {
      throw new Error("Сервер вернул некорректный список участников.");
    }
    return response.items;
  },

  create(input: UserInput): Promise<User> {
    return request<User>("/api/users", "POST", input);
  },

  update(id: number, input: UserInput): Promise<User> {
    return request<User>(`/api/users/${id}`, "PUT", input);
  },

  remove(id: number): Promise<void> {
    return request<void>(`/api/users/${id}`, "DELETE");
  },
};
