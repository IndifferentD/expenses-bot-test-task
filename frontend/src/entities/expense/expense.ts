export const categoryOptions = [
  { value: 'food', label: 'Еда и продукты' },
  { value: 'transport', label: 'Транспорт' },
  { value: 'housing', label: 'Жильё и коммунальные услуги' },
  { value: 'entertainment', label: 'Развлечения' },
  { value: 'health', label: 'Здоровье' },
  { value: 'shopping', label: 'Покупки' },
  { value: 'other', label: 'Другое' },
] as const;

export type Category = (typeof categoryOptions)[number]['value'];

export interface Expense {
  id: number;
  user_id: number;
  amount: number;
  category: Category;
  spent_at: string;
  comment: string | null;
  created_at: string;
  updated_at: string;
}

export type ExpenseInput = Pick<Expense, 'user_id' | 'amount' | 'category' | 'spent_at' | 'comment'>;

export function categoryLabel(category: string): string {
  return categoryOptions.find((option) => option.value === category)?.label ?? category;
}

export function formatAmount(amount: number): string {
  return new Intl.NumberFormat('ru-RU').format(amount);
}

export function formatSpentAt(value: string): string {
  const [year, month, day] = value.split('-');
  return year && month && day ? `${day}.${month}.${year}` : value;
}

export function formatTimestamp(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat('ru-RU', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(date);
}
