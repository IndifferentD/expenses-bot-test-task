<script setup lang="ts">
import { onMounted, ref } from 'vue';

defineProps<{ message: string }>();
const emit = defineEmits<{ close: [] }>();
const dialog = ref<HTMLDialogElement | null>(null);

onMounted(() => dialog.value?.showModal());

function closeOnBackdrop(event: MouseEvent): void {
  const element = dialog.value;
  if (!element || event.target !== element) return;
  const rect = element.getBoundingClientRect();
  if (event.clientX < rect.left || event.clientX > rect.right || event.clientY < rect.top || event.clientY > rect.bottom) {
    element.close();
  }
}
</script>

<template>
  <Teleport to="body">
    <dialog ref="dialog" class="error-dialog" aria-labelledby="error-dialog-title" @click="closeOnBackdrop" @close="emit('close')">
      <div class="dialog-icon" aria-hidden="true">!</div>
      <h2 id="error-dialog-title">Не удалось выполнить действие</h2>
      <p>{{ message }}</p>
      <form method="dialog">
        <button class="button button--primary" type="submit" autofocus>ОК</button>
      </form>
    </dialog>
  </Teleport>
</template>
