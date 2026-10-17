<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { expensesApi, type ExpenseFilters } from '../../entities/expense/expensesApi';
import CategoryIcon from '../../entities/expense/CategoryIcon.vue';
import ExpenseRow from '../../entities/expense/ExpenseRow.vue';
import { categoryOptions, formatAmount, type Category, type Expense, type ExpenseInput } from '../../entities/expense/expense';
import { usersApi } from '../../entities/user/usersApi';
import type { User } from '../../entities/user/user';
import ErrorDialog from '../../shared/ErrorDialog.vue';

const expenses = ref<Expense[]>([]);
const users = ref<User[]>([]);
const expensesLoading = ref(true);
const usersLoading = ref(true);
const expensesLoadError = ref<string | null>(null);
const usersLoadError = ref<string | null>(null);
const errorMessage = ref<string | null>(null);
const adding = ref(false);
const editingId = ref<number | null>(null);
const savingKey = ref<number | 'new' | null>(null);
const deletingId = ref<number | null>(null);
const nextCursor = ref<string | null>(null);
const loadingMore = ref(false);

const selectedUserIds = ref<number[]>([]);
const selectedCategories = ref<Category[]>([]);
const dateFrom = ref('');
const dateTo = ref('');

const isLocked = computed(() => adding.value || editingId.value !== null || savingKey.value !== null || deletingId.value !== null);
const loading = computed(() => expensesLoading.value || usersLoading.value);
const loadError = computed(() => expensesLoadError.value || usersLoadError.value);
const hasFilters = computed(() => Boolean(selectedUserIds.value.length || selectedCategories.value.length || dateFrom.value || dateTo.value));
const invalidPeriod = computed(() => Boolean(dateFrom.value && dateTo.value && dateFrom.value > dateTo.value));
const userFilterLabel = computed(() => {
  if (selectedUserIds.value.length === 0) return 'Все участники';
  if (selectedUserIds.value.length === 1) {
    return users.value.find((user) => user.id === selectedUserIds.value[0])?.alias ?? '1 участник';
  }
  return `Выбрано: ${selectedUserIds.value.length}`;
});
const categoryFilterLabel = computed(() => {
  if (selectedCategories.value.length === 0) return 'Все категории';
  if (selectedCategories.value.length === 1) {
    return categoryOptions.find((option) => option.value === selectedCategories.value[0])?.label ?? '1 категория';
  }
  return `Выбрано: ${selectedCategories.value.length}`;
});
const selectedCategory = computed(() => selectedCategories.value.length === 1 ? selectedCategories.value[0] : null);

function matchesFilters(expense: Expense): boolean {
  return (selectedUserIds.value.length === 0 || selectedUserIds.value.includes(expense.user_id))
    && (selectedCategories.value.length === 0 || selectedCategories.value.includes(expense.category))
    && (!dateFrom.value || expense.spent_at >= dateFrom.value)
    && (!dateTo.value || expense.spent_at <= dateTo.value);
}

const visibleExpenses = computed(() => invalidPeriod.value ? [] : expenses.value);
const visibleAmount = computed(() => visibleExpenses.value.reduce((sum, expense) => sum + expense.amount, 0));

function messageFrom(error: unknown): string {
  return error instanceof Error ? error.message : 'Произошла неизвестная ошибка.';
}

let latestExpenseRequest = 0;

function currentFilters(): ExpenseFilters {
  return {
    userIds: selectedUserIds.value,
    categories: selectedCategories.value,
    dateFrom: dateFrom.value,
    dateTo: dateTo.value,
  };
}

async function loadExpenses(): Promise<void> {
  const requestId = ++latestExpenseRequest;
  expensesLoading.value = true;
  expensesLoadError.value = null;
  nextCursor.value = null;
  if (invalidPeriod.value) {
    expenses.value = [];
    expensesLoading.value = false;
    return;
  }
  try {
    const page = await expensesApi.list(currentFilters());
    if (requestId !== latestExpenseRequest) return;
    expenses.value = page.items;
    nextCursor.value = page.next_cursor;
  } catch (error) {
    if (requestId === latestExpenseRequest) expensesLoadError.value = messageFrom(error);
  } finally {
    if (requestId === latestExpenseRequest) expensesLoading.value = false;
  }
}

async function loadMoreExpenses(): Promise<void> {
  if (!nextCursor.value || loading.value || loadingMore.value || isLocked.value) return;
  const requestId = latestExpenseRequest;
  const cursor = nextCursor.value;
  loadingMore.value = true;
  try {
    const page = await expensesApi.list(currentFilters(), cursor);
    if (requestId !== latestExpenseRequest) return;
    expenses.value = [...expenses.value, ...page.items];
    nextCursor.value = page.next_cursor;
  } catch (error) {
    if (requestId === latestExpenseRequest) errorMessage.value = messageFrom(error);
  } finally {
    loadingMore.value = false;
  }
}

async function loadUsers(): Promise<void> {
  usersLoading.value = true;
  usersLoadError.value = null;
  try {
    users.value = await usersApi.list();
  } catch (error) {
    usersLoadError.value = messageFrom(error);
  } finally {
    usersLoading.value = false;
  }
}

async function load(): Promise<void> {
  await Promise.all([loadExpenses(), loadUsers()]);
}

function clearFilters(): void {
  selectedUserIds.value = [];
  selectedCategories.value = [];
  dateFrom.value = '';
  dateTo.value = '';
}

function startAdd(): void {
  if (!isLocked.value && users.value.length > 0) adding.value = true;
}

function startEdit(id: number): void {
  if (!isLocked.value) editingId.value = id;
}

async function saveNew(input: ExpenseInput): Promise<void> {
  if (savingKey.value !== null) return;
  savingKey.value = 'new';
  try {
    const created = await expensesApi.create(input);
    if (matchesFilters(created)) void loadExpenses();
    else clearFilters();
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
    if (matchesFilters(updated)) void loadExpenses();
    else clearFilters();
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
    void loadExpenses();
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    deletingId.value = null;
  }
}

watch([selectedUserIds, selectedCategories, dateFrom, dateTo], () => { void loadExpenses(); }, { deep: true });
onMounted(() => { void load(); });
</script>

<template>
  <main class="page">
    <div class="page-intro">
      <div>
        <h1 class="eyebrow"><span class="eyebrow__dot" aria-hidden="true"></span> Траты семьи</h1>
        <p class="page-intro__subtitle">Все расходы в одном месте — добавляйте, находите и редактируйте записи прямо в таблице.</p>
      </div>
      <div class="summary" aria-label="Сводка расходов">
        <div class="summary__item">
          <span>Загружено записей</span>
          <strong>{{ formatAmount(visibleExpenses.length) }}</strong>
        </div>
        <div class="summary__divider" aria-hidden="true"></div>
        <div class="summary__item">
          <span>Сумма загруженных</span>
          <strong>{{ formatAmount(visibleAmount) }}</strong>
        </div>
      </div>
    </div>

    <section class="expenses-panel" aria-labelledby="expenses-title">
      <div class="panel-heading">
        <div class="panel-heading__left">
          <button class="button button--primary add-button" type="button" :disabled="loading || Boolean(loadError) || isLocked || users.length === 0" @click="startAdd">
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
          <span class="filter-label">Участник семьи</span>
          <details class="multi-select">
            <summary class="filter-control multi-select__trigger" :aria-disabled="isLocked" :tabindex="isLocked ? -1 : 0" @click="isLocked && $event.preventDefault()">
              {{ userFilterLabel }}
            </summary>
            <div class="multi-select__menu">
              <label v-for="user in users" :key="user.id" class="multi-select__option">
                <input v-model="selectedUserIds" type="checkbox" :value="user.id" :disabled="isLocked" />
                <span>{{ user.alias }}</span>
              </label>
              <span v-if="users.length === 0" class="multi-select__empty">Участников пока нет</span>
            </div>
          </details>
        </div>
        <div class="filter-field">
          <span class="filter-label">Категория</span>
          <details class="multi-select">
            <summary class="filter-control multi-select__trigger" :aria-disabled="isLocked" :tabindex="isLocked ? -1 : 0" @click="isLocked && $event.preventDefault()">
              <span v-if="selectedCategory" class="category-choice">
                <CategoryIcon :category="selectedCategory" />
                <span class="category-text">{{ categoryFilterLabel }}</span>
              </span>
              <span v-else>{{ categoryFilterLabel }}</span>
            </summary>
            <div class="multi-select__menu">
              <label v-for="option in categoryOptions" :key="option.value" class="multi-select__option">
                <input v-model="selectedCategories" type="checkbox" :value="option.value" :disabled="isLocked" />
                <span class="category-choice">
                  <CategoryIcon :category="option.value" />
                  <span>{{ option.label }}</span>
                </span>
              </label>
            </div>
          </details>
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
      <p v-if="!loading && !loadError && users.length === 0" class="filter-warning">Чтобы добавить трату, сначала добавьте участника на вкладке «Семья».</p>

      <div class="table-scroll">
        <table class="expenses-table">
          <caption class="sr-only">Траты членов семьи</caption>
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">Участник</th>
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
              :users="users"
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
              :users="users"
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
        <span>Показано трат: {{ visibleExpenses.length }}</span>
        <button v-if="nextCursor && !loadError" class="button button--primary" type="button" :disabled="loading || loadingMore || isLocked" @click="loadMoreExpenses">
          {{ loadingMore ? 'Загружаем…' : 'Показать ещё' }}
        </button>
      </div>
    </section>

    <ErrorDialog v-if="errorMessage" :message="errorMessage" @close="errorMessage = null" />
  </main>
</template>
