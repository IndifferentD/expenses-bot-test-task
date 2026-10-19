export interface User {
  id: number;
  telegram_user_id: number;
  alias: string;
}

export type UserInput = Pick<User, "telegram_user_id" | "alias">;
