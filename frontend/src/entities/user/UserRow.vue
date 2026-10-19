<script setup lang="ts">
import { nextTick, onMounted, reactive, ref, watch } from "vue";
import type { User, UserInput } from "./user";

const props = defineProps<{
  user: User | null;
  editing: boolean;
  saving: boolean;
  locked: boolean;
}>();

const emit = defineEmits<{
  edit: [];
  cancel: [];
  save: [input: UserInput];
  delete: [];
  invalid: [message: string];
}>();

function makeDraft(user: User | null) {
  return {
    telegramUserId: user ? String(user.telegram_user_id) : "",
    alias: user?.alias ?? "",
  };
}

const draft = reactive(makeDraft(props.user));
const firstField = ref<HTMLInputElement | null>(null);

watch(
  () => props.editing,
  async (editing) => {
    if (editing) {
      Object.assign(draft, makeDraft(props.user));
      await nextTick();
      firstField.value?.focus();
    }
  },
);

onMounted(() => {
  if (props.editing) firstField.value?.focus();
});

function submit(): void {
  const telegramUserId = Number(draft.telegramUserId);
  const alias = draft.alias.trim();
  if (!Number.isSafeInteger(telegramUserId) || telegramUserId <= 0) {
    emit("invalid", "Telegram ID должен быть положительным целым числом.");
    return;
  }
  if (!alias) {
    emit("invalid", "Укажите имя участника.");
    return;
  }
  emit("save", { telegram_user_id: telegramUserId, alias });
}
</script>

<template>
  <tr
    :class="{ 'user-row--editing': editing, 'user-row--new': !user }"
    :aria-busy="saving"
  >
    <th scope="row" class="cell-id">{{ user ? `#${user.id}` : "Новый" }}</th>
    <td>
      <input
        v-if="editing"
        ref="firstField"
        v-model="draft.telegramUserId"
        class="table-input table-input--user"
        type="number"
        min="1"
        step="1"
        inputmode="numeric"
        aria-label="Telegram ID участника"
        placeholder="Telegram ID"
        :disabled="saving"
        @keydown.enter.prevent="submit"
      />
      <span v-else>{{ user?.telegram_user_id }}</span>
    </td>
    <td>
      <input
        v-if="editing"
        v-model="draft.alias"
        class="table-input"
        type="text"
        aria-label="Имя участника"
        placeholder="Имя участника"
        :disabled="saving"
        @keydown.enter.prevent="submit"
      />
      <strong v-else>{{ user?.alias }}</strong>
    </td>
    <td class="cell-actions">
      <div class="row-actions">
        <template v-if="editing">
          <button
            class="icon-button icon-button--save"
            type="button"
            :aria-label="
              user
                ? `Сохранить изменения участника #${user.id}`
                : 'Добавить участника'
            "
            :title="user ? 'Сохранить' : 'Добавить'"
            :disabled="saving"
            @click="submit"
          >
            <svg
              v-if="user"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="m5 12 4.5 4.5L19 7" />
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              aria-hidden="true"
            >
              <path d="M12 5v14M5 12h14" />
            </svg>
          </button>
          <button
            class="icon-button icon-button--cancel"
            type="button"
            :aria-label="
              user
                ? `Отменить изменения участника #${user.id}`
                : 'Отменить добавление участника'
            "
            title="Отменить"
            :disabled="saving"
            @click="emit('cancel')"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <circle cx="12" cy="12" r="9" />
              <path d="M5.6 5.6 18.4 18.4" />
            </svg>
          </button>
        </template>
        <template v-else>
          <button
            class="icon-button icon-button--edit"
            type="button"
            :aria-label="`Редактировать участника #${user?.id}`"
            title="Редактировать"
            :disabled="locked"
            @click="emit('edit')"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path
                d="m15.5 5.5 3 3M4 20l4.3-1 10.9-10.9a2.1 2.1 0 0 0-3-3L5.3 16 4 20Z"
              />
            </svg>
          </button>
          <button
            class="icon-button icon-button--delete"
            type="button"
            :aria-label="`Удалить участника #${user?.id}`"
            title="Удалить"
            :disabled="locked"
            @click="emit('delete')"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              aria-hidden="true"
            >
              <path d="M6 6 18 18M18 6 6 18" />
            </svg>
          </button>
        </template>
      </div>
    </td>
  </tr>
</template>
