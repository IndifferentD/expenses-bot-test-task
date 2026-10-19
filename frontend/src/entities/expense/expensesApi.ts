import { request } from "../../shared/http";
import type { Category, Expense, ExpenseInput } from "./expense";

export interface ExpensePage {
  items: Expense[];
  next_cursor: string | null;
}

export interface ExpenseFilters {
  userIds: number[];
  categories: Category[];
  dateFrom: string;
  dateTo: string;
}

export const expensesApi = {
  async list(filters: ExpenseFilters, cursor?: string): Promise<ExpensePage> {
    const params = new URLSearchParams();
    if (filters.userIds.length)
      params.set("user_id", filters.userIds.join(","));
    if (filters.categories.length)
      params.set("category", filters.categories.join(","));
    if (filters.dateFrom) params.set("date_from", filters.dateFrom);
    if (filters.dateTo) params.set("date_to", filters.dateTo);
    if (cursor) params.set("cursor", cursor);
    const query = params.toString();
    const response = await request<ExpensePage>(
      `/api/expenses${query ? `?${query}` : ""}`,
    );
    if (
      !Array.isArray(response.items) ||
      (response.next_cursor !== null &&
        typeof response.next_cursor !== "string")
    ) {
      throw new Error("Сервер вернул некорректный список трат.");
    }
    return response;
  },

  create(input: ExpenseInput): Promise<Expense> {
    return request<Expense>("/api/expenses", "POST", input);
  },

  update(id: number, input: ExpenseInput): Promise<Expense> {
    return request<Expense>(`/api/expenses/${id}`, "PUT", input);
  },

  remove(id: number): Promise<void> {
    return request<void>(`/api/expenses/${id}`, "DELETE");
  },
};
