<script setup lang="ts">
import { nextTick, onMounted, reactive, ref, watch } from 'vue';
import type { User } from '../user/user';
import {
  categoryLabel,
  categoryOptions,
  formatAmount,
  formatSpentAt,
  formatTimestamp,
  type Category,
  type Expense,
  type ExpenseInput,
} from './expense';

const props = defineProps<{
  expense: Expense | null;
  editing: boolean;
  saving: boolean;
  locked: boolean;
  users: User[];
}>();

const emit = defineEmits<{
  edit: [];
  cancel: [];
  save: [input: ExpenseInput];
  delete: [];
  invalid: [message: string];
}>();

interface Draft {
  userId: string;
  amount: string;
  category: Category | '';
  spentAt: string;
  comment: string;
}

function today(): string {
  const date = new Date();
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function makeDraft(expense: Expense | null): Draft {
  return {
    userId: expense ? String(expense.user_id) : '',
    amount: expense ? String(expense.amount) : '',
    category: expense?.category ?? '',
    spentAt: expense?.spent_at ?? today(),
    comment: expense?.comment ?? '',
  };
}

const draft = reactive(makeDraft(props.expense));
const firstField = ref<HTMLSelectElement | null>(null);

watch(() => props.editing, async (editing) => {
  if (editing) {
    Object.assign(draft, makeDraft(props.expense));
    await nextTick();
    firstField.value?.focus();
  }
});

onMounted(() => {
  if (props.editing) firstField.value?.focus();
});

function submit(): void {
  const userId = Number(draft.userId);
  const amount = Number(draft.amount);

  if (!Number.isSafeInteger(userId) || userId <= 0) {
    emit('invalid', 'Выберите участника семьи.');
    return;
  }
  if (!props.users.some((user) => user.id === userId)) {
    emit('invalid', 'Выбранный участник не найден.');
    return;
  }
  if (!Number.isSafeInteger(amount) || amount <= 0) {
    emit('invalid', 'Сумма должна быть положительным целым числом.');
    return;
  }
  if (!draft.category) {
    emit('invalid', 'Выберите категорию траты.');
    return;
  }
  if (!/^\d{4}-\d{2}-\d{2}$/.test(draft.spentAt)) {
    emit('invalid', 'Укажите дату траты.');
    return;
  }

  emit('save', {
    user_id: userId,
    amount,
    category: draft.category,
    spent_at: draft.spentAt,
    comment: draft.comment.trim() || null,
  });
}
</script>

<template>
  <tr :class="{ 'expense-row--editing': editing, 'expense-row--new': !expense }" :aria-busy="saving">
    <th scope="row" class="cell-id">{{ expense ? `#${expense.id}` : 'Новая' }}</th>
    <td>
      <select
        v-if="editing"
        ref="firstField"
        v-model="draft.userId"
        class="table-input table-input--user"
        aria-label="Участник семьи"
        :disabled="saving"
      >
        <option value="" disabled>Выберите участника</option>
        <option v-for="user in users" :key="user.id" :value="String(user.id)">{{ user.alias }}</option>
      </select>
      <span v-else>{{ users.find((user) => user.id === expense?.user_id)?.alias ?? '—' }}</span>
    </td>
    <td class="cell-amount">
      <input
        v-if="editing"
        v-model="draft.amount"
        class="table-input table-input--amount"
        type="number"
        min="1"
        step="1"
        inputmode="numeric"
        aria-label="Сумма траты"
        placeholder="Сумма"
        :disabled="saving"
      />
      <strong v-else-if="expense">{{ formatAmount(expense.amount) }}</strong>
    </td>
    <td>
      <select v-if="editing" v-model="draft.category" class="table-input table-select" aria-label="Категория траты" :disabled="saving">
        <option value="" disabled>Выберите категорию</option>
        <option v-for="option in categoryOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
      <span v-else-if="expense" class="category-pill">{{ categoryLabel(expense.category) }}</span>
    </td>
    <td>
      <input
        v-if="editing"
        v-model="draft.spentAt"
        class="table-input table-input--date"
        type="date"
        aria-label="Дата траты"
        :disabled="saving"
      />
      <span v-else-if="expense">{{ formatSpentAt(expense.spent_at) }}</span>
    </td>
    <td class="cell-comment">
      <input
        v-if="editing"
        v-model="draft.comment"
        class="table-input table-input--comment"
        type="text"
        aria-label="Комментарий"
        placeholder="Необязательно"
        :disabled="saving"
      />
      <span v-else :class="{ 'text-muted': !expense?.comment }">{{ expense?.comment || '—' }}</span>
    </td>
    <td class="cell-timestamp">{{ expense ? formatTimestamp(expense.created_at) : '—' }}</td>
    <td class="cell-timestamp">{{ expense ? formatTimestamp(expense.updated_at) : '—' }}</td>
    <td class="cell-actions">
      <div class="row-actions">
        <template v-if="editing">
          <button
            class="icon-button icon-button--save"
            type="button"
            :aria-label="expense ? `Сохранить изменения траты #${expense.id}` : 'Добавить трату'"
            :title="expense ? 'Сохранить' : 'Добавить'"
            :disabled="saving"
            @click="submit"
          >
            <svg v-if="expense" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="m5 12 4.5 4.5L19 7" />
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
              <path d="M12 5v14M5 12h14" />
            </svg>
          </button>
          <button
            class="icon-button icon-button--cancel"
            type="button"
            :aria-label="expense ? `Отменить изменения траты #${expense.id}` : 'Отменить добавление траты'"
            title="Отменить"
            :disabled="saving"
            @click="emit('cancel')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="12" cy="12" r="9" />
              <path d="M5.6 5.6 18.4 18.4" />
            </svg>
          </button>
        </template>
        <template v-else>
          <button
            class="icon-button icon-button--edit"
            type="button"
            :aria-label="`Редактировать трату #${expense?.id}`"
            title="Редактировать"
            :disabled="locked"
            @click="emit('edit')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="m15.5 5.5 3 3M4 20l4.3-1 10.9-10.9a2.1 2.1 0 0 0-3-3L5.3 16 4 20Z" />
            </svg>
          </button>
          <button
            class="icon-button icon-button--delete"
            type="button"
            :aria-label="`Удалить трату #${expense?.id}`"
            title="Удалить"
            :disabled="locked"
            @click="emit('delete')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
              <path d="M6 6 18 18M18 6 6 18" />
            </svg>
          </button>
        </template>
      </div>
    </td>
  </tr>
</template>
