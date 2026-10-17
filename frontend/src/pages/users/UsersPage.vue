<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import UserRow from '../../entities/user/UserRow.vue';
import { usersApi } from '../../entities/user/usersApi';
import type { User, UserInput } from '../../entities/user/user';
import ErrorDialog from '../../shared/ErrorDialog.vue';

const users = ref<User[]>([]);
const loading = ref(true);
const loadError = ref<string | null>(null);
const errorMessage = ref<string | null>(null);
const adding = ref(false);
const editingId = ref<number | null>(null);
const savingKey = ref<number | 'new' | null>(null);
const deletingId = ref<number | null>(null);

const isLocked = computed(() => adding.value || editingId.value !== null || savingKey.value !== null || deletingId.value !== null);

function messageFrom(error: unknown): string {
  return error instanceof Error ? error.message : 'Произошла неизвестная ошибка.';
}

async function load(): Promise<void> {
  loading.value = true;
  loadError.value = null;
  try {
    users.value = await usersApi.list();
  } catch (error) {
    loadError.value = messageFrom(error);
  } finally {
    loading.value = false;
  }
}

function startAdd(): void {
  if (!isLocked.value) adding.value = true;
}

function startEdit(id: number): void {
  if (!isLocked.value) editingId.value = id;
}

async function saveNew(input: UserInput): Promise<void> {
  if (savingKey.value !== null) return;
  savingKey.value = 'new';
  try {
    const created = await usersApi.create(input);
    users.value = [...users.value, created].sort((a, b) => a.alias.localeCompare(b.alias, 'ru') || a.id - b.id);
    adding.value = false;
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    savingKey.value = null;
  }
}

async function saveExisting(id: number, input: UserInput): Promise<void> {
  if (savingKey.value !== null) return;
  savingKey.value = id;
  try {
    const updated = await usersApi.update(id, input);
    users.value = users.value.map((user) => user.id === id ? updated : user)
      .sort((a, b) => a.alias.localeCompare(b.alias, 'ru') || a.id - b.id);
    editingId.value = null;
  } catch (error) {
    errorMessage.value = messageFrom(error);
  } finally {
    savingKey.value = null;
  }
}

async function remove(user: User): Promise<void> {
  if (isLocked.value || !window.confirm(`Удалить участника «${user.alias}»? Это действие нельзя отменить.`)) return;
  deletingId.value = user.id;
  try {
    await usersApi.remove(user.id);
    users.value = users.value.filter((item) => item.id !== user.id);
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
        <h1 class="eyebrow"><span class="eyebrow__dot" aria-hidden="true"></span> Участники</h1>
        <p class="page-intro__subtitle">Добавьте членов семьи и укажите их Telegram ID, чтобы связывать с ними траты.</p>
      </div>
    </div>

    <section class="users-panel" aria-labelledby="users-title">
      <div class="panel-heading">
        <div class="panel-heading__left">
          <button class="button button--primary add-button" type="button" :disabled="loading || Boolean(loadError) || isLocked" @click="startAdd">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
              <path d="M12 5v14M5 12h14" />
            </svg>
            Добавить участника
          </button>
          <div>
            <h2 id="users-title">Члены семьи</h2>
            <p>Участника с тратами удалить нельзя</p>
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

      <div class="table-scroll">
        <table class="users-table">
          <caption class="sr-only">Члены семьи</caption>
          <thead>
            <tr>
              <th scope="col">ID</th>
              <th scope="col">Telegram ID</th>
              <th scope="col">Имя</th>
              <th scope="col" class="actions-heading">Действия</th>
            </tr>
          </thead>
          <tbody>
            <UserRow
              v-if="adding"
              :user="null"
              :editing="true"
              :saving="savingKey === 'new'"
              :locked="false"
              @cancel="adding = false"
              @save="saveNew"
              @invalid="errorMessage = $event"
            />
            <tr v-if="loading">
              <td colspan="4" class="state-cell">Загружаем участников…</td>
            </tr>
            <tr v-else-if="loadError">
              <td colspan="4" class="state-cell state-cell--error">
                <span>{{ loadError }}</span>
                <button class="button button--quiet" type="button" @click="load">Повторить</button>
              </td>
            </tr>
            <tr v-else-if="users.length === 0 && !adding">
              <td colspan="4" class="state-cell">Участников пока нет. Добавьте первого.</td>
            </tr>
            <UserRow
              v-for="user in users"
              v-show="!loading && !loadError"
              :key="user.id"
              :user="user"
              :editing="editingId === user.id"
              :saving="savingKey === user.id"
              :locked="isLocked"
              @edit="startEdit(user.id)"
              @cancel="editingId = null"
              @save="(input) => saveExisting(user.id, input)"
              @delete="remove(user)"
              @invalid="errorMessage = $event"
            />
          </tbody>
        </table>
      </div>

      <div class="panel-footer"><span>Всего участников: {{ users.length }}</span></div>
    </section>

    <ErrorDialog v-if="errorMessage" :message="errorMessage" @close="errorMessage = null" />
  </main>
</template>
