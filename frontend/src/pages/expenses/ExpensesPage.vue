<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { expensesApi } from '../../entities/expense/expensesApi';
import ExpenseRow from '../../entities/expense/ExpenseRow.vue';
import { categoryOptions, formatAmount, type Category, type Expense, type ExpenseInput } from '../../entities/expense/expense';
import ErrorDialog from '../../shared/ErrorDialog.vue';

const expenses = ref<Expense[]>([]);
const loading = ref(true);
const loadError = ref<string | null>(null);
const errorMessage = ref<string | null>(null);
const adding = ref(false);
const editingId = ref<number | null>(null);
const savingKey = ref<number | 'new' | null>(null);
const deletingId = ref<number | null>(null);

const userFilter = ref('');
const categoryFilter = ref<Category | ''>('');
const dateFrom = ref('');
const dateTo = ref('');

const isLocked = computed(() => adding.value || editingId.value !== null || savingKey.value !== null || deletingId.value !== null);
const hasFilters = computed(() => Boolean(userFilter.value || categoryFilter.value || dateFrom.value || dateTo.value));
const invalidPeriod = computed(() => Boolean(dateFrom.value && dateTo.value && dateFrom.value > dateTo.value));
const users = computed(() => [...new Set(expenses.value.map((expense) => expense.user_id))].sort((a, b) => a - b));

function matchesFilters(expense: Expense): boolean {
  return (!userFilter.value || String(expense.user_id) === userFilter.value)
    && (!categoryFilter.value || expense.category === categoryFilter.value)
    && (!dateFrom.value || expense.spent_at >= dateFrom.value)
    && (!dateTo.value || expense.spent_at <= dateTo.value);
}

const visibleExpenses = computed(() => invalidPeriod.value ? [] : expenses.value.filter(matchesFilters));
const visibleAmount = computed(() => visibleExpenses.value.reduce((sum, expense) => sum + expense.amount, 0));

function sortExpenses(items: Expense[]): Expense[] {
  return [...items].sort((a, b) => b.spent_at.localeCompare(a.spent_at) || b.id - a.id);
}

function messageFrom(error: unknown): string {
  return error instanceof Error ? error.message : 'Произошла неизвестная ошибка.';
}

async function load(): Promise<void> {
  loading.value = true;
  loadError.value = null;
  try {
    expenses.value = sortExpenses(await expensesApi.list());
  } catch (error) {
    loadError.value = messageFrom(error);
  } finally {
    loading.value = false;
  }
}

function clearFilters(): void {
  userFilter.value = '';
  categoryFilter.value = '';
  dateFrom.value = '';
  dateTo.value = '';
}

function startAdd(): void {
  if (!isLocked.value) adding.value = true;
}

function startEdit(id: number): void {
  if (!isLocked.value) editingId.value = id;
}

async function saveNew(input: ExpenseInput): Promise<void> {
  if (savingKey.value !== null) return;
  savingKey.value = 'new';
  try {
    const created = await expensesApi.create(input);
    expenses.value = sortExpenses([...expenses.value, created]);
    if (!matchesFilters(created)) clearFilters();
    adding.value = false;
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    savingKey.value = null;
  }
}

async function saveExisting(id: number, input: ExpenseInput): Promise<void> {
  if (savingKey.value !== null) return;
  savingKey.value = id;
  try {
    const updated = await expensesApi.update(id, input);
    expenses.value = sortExpenses(expenses.value.map((expense) => expense.id === id ? updated : expense));
    if (!matchesFilters(updated)) clearFilters();
    editingId.value = null;
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    savingKey.value = null;
  }
}

async function remove(expense: Expense): Promise<void> {
  if (isLocked.value || !window.confirm(`Удалить трату #${expense.id}? Это действие нельзя отменить.`)) return;
  deletingId.value = expense.id;
  try {
    await expensesApi.remove(expense.id);
    expenses.value = expenses.value.filter((item) => item.id !== expense.id);
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    deletingId.value = null;
  }
}

onMounted(() => { void load(); });
</script>

<template>
  <main class="page">
    <div class="page-intro">
      <div>
        <p class="eyebrow"><span class="eyebrow__dot" aria-hidden="true"></span> Общая картина</p>
        <h1>Траты семьи</h1>
        <p class="page-intro__subtitle">Все расходы в одном месте — добавляйте, находите и редактируйте записи прямо в таблице.</p>
      </div>
      <div class="summary" aria-label="Сводка расходов">
        <div class="summary__item">
          <span>Записей</span>
          <strong>{{ formatAmount(visibleExpenses.length) }}</strong>
        </div>
        <div class="summary__divider" aria-hidden="true"></div>
        <div class="summary__item">
          <span>Сумма</span>
          <strong>{{ formatAmount(visibleAmount) }}</strong>
        </div>
      </div>
    </div>

    <section class="expenses-panel" aria-labelledby="expenses-title">
      <div class="panel-heading">
        <div class="panel-heading__left">
          <button class="button button--primary add-button" type="button" :disabled="loading || Boolean(loadError) || isLocked" @click="startAdd">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
              <path d="M12 5v14M5 12h14" />
            </svg>
            Добавить трату
          </button>
          <div>
            <h2 id="expenses-title">Список трат</h2>
            <p>Редактирование и удаление доступны в каждой строке</p>
          </div>
        </div>
        <button class="button button--quiet" type="button" :disabled="loading || isLocked" @click="load">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M20 7v5h-5M4 17v-5h5" />
            <path d="M5.7 9A7 7 0 0 1 18 7l2 5M4 12l2 5a7 7 0 0 0 12.3-2" />
          </svg>
          Обновить
        </button>
      </div>

      <div class="filters" aria-label="Фильтры трат">
        <div class="filter-field">
          <label for="filter-user">Участник семьи</label>
          <select id="filter-user" v-model="userFilter" class="filter-control" :disabled="isLocked">
            <option value="">Все участники</option>
            <option v-for="userId in users" :key="userId" :value="String(userId)">Telegram ID {{ userId }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label for="filter-category">Категория</label>
          <select id="filter-category" v-model="categoryFilter" class="filter-control" :disabled="isLocked">
            <option value="">Все категории</option>
            <option v-for="option in categoryOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
          </select>
        </div>
        <div class="filter-field">
          <label for="filter-from">Дата с</label>
          <input id="filter-from" v-model="dateFrom" class="filter-control" type="date" :max="dateTo || undefined" :disabled="isLocked" />
        </div>
        <div class="filter-field">
          <label for="filter-to">Дата по</label>
          <input id="filter-to" v-model="dateTo" class="filter-control" type="date" :min="dateFrom || undefined" :disabled="isLocked" />
        </div>
        <button class="button button--clear" type="button" :disabled="!hasFilters || isLocked" @click="clearFilters">Сбросить</button>
      </div>
      <p v-if="invalidPeriod" class="filter-warning" role="alert">Дата «с» должна быть не позже даты «по».</p>

      <div class="table-scroll">
        <table class="expenses-table">
          <caption class="sr-only">Траты членов семьи</caption>
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">Telegram ID</th>
              <th scope="col">Сумма</th>
              <th scope="col">Категория</th>
              <th scope="col">Дата траты</th>
              <th scope="col">Комментарий</th>
              <th scope="col">Создана</th>
              <th scope="col">Обновлена</th>
              <th scope="col" class="actions-heading">Действия</th>
            </tr>
          </thead>
          <tbody>
            <ExpenseRow
              v-if="adding"
              :expense="null"
              :editing="true"
              :saving="savingKey === 'new'"
              :locked="false"
              @cancel="adding = false"
              @save="saveNew"
              @invalid="errorMessage = $event"
            />
            <tr v-if="loading">
              <td colspan="9" class="state-cell">Загружаем траты…</td>
            </tr>
            <tr v-else-if="loadError">
              <td colspan="9" class="state-cell state-cell--error">
                <span>{{ loadError }}</span>
                <button class="button button--quiet" type="button" @click="load">Повторить</button>
              </td>
            </tr>
            <tr v-else-if="visibleExpenses.length === 0 && !adding">
              <td colspan="9" class="state-cell">{{ hasFilters ? 'По выбранным фильтрам трат не найдено.' : 'Трат пока нет. Добавьте первую запись.' }}</td>
            </tr>
            <ExpenseRow
              v-for="expense in visibleExpenses"
              v-show="!loading && !loadError"
              :key="expense.id"
              :expense="expense"
              :editing="editingId === expense.id"
              :saving="savingKey === expense.id"
              :locked="isLocked"
              @edit="startEdit(expense.id)"
              @cancel="editingId = null"
              @save="(input) => saveExisting(expense.id, input)"
              @delete="remove(expense)"
              @invalid="errorMessage = $event"
            />
          </tbody>
        </table>
      </div>

      <div class="panel-footer">
        <span>Показано {{ visibleExpenses.length }} из {{ expenses.length }}</span>
      </div>
    </section>

    <ErrorDialog v-if="errorMessage" :message="errorMessage" @close="errorMessage = null" />
  </main>
</template>
