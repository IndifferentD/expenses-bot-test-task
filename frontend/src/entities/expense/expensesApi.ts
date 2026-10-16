import { request } from '../../shared/http';
import type { Expense, ExpenseInput } from './expense';

interface ExpenseListResponse {
  items: Expense[];
}

export const expensesApi = {
  async list(): Promise<Expense[]> {
    const response = await request<ExpenseListResponse>('/api/expenses');
    if (!Array.isArray(response.items)) {
      throw new Error('Сервер вернул некорректный список трат.');
    }
    return response.items;
  },

  create(input: ExpenseInput): Promise<Expense> {
    return request<Expense>('/api/expenses', 'POST', input);
  },

  update(id: number, input: ExpenseInput): Promise<Expense> {
    return request<Expense>(`/api/expenses/${id}`, 'PUT', input);
  },

  remove(id: number): Promise<void> {
    return request<void>(`/api/expenses/${id}`, 'DELETE');
  },
};
