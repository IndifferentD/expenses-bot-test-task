<script setup lang="ts">
import type { Category } from './expense';

defineProps<{ category: Category }>();
</script>

<template>
  <svg
    class="category-icon"
    :class="`category-icon--${category}`"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.8"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    focusable="false"
  >
    <template v-if="category === 'food'">
      <circle cx="12" cy="12" r="5" />
      <path d="M3 3v7m3-7v7M3 7h3m-1.5 3v11M20 3c-2 2-3 5-3 9h3v9" />
    </template>
    <template v-else-if="category === 'transport'">
      <path d="m5 16 1.5-7h11L19 16M5 16h14v3H5zM7 9l1-4h8l1 4M7 19v2m10-2v2" />
      <path d="M8 13h.01M16 13h.01" stroke-width="2.8" />
    </template>
    <template v-else-if="category === 'housing'">
      <path d="m3 10 9-7 9 7v11H3zM9 21v-7h6v7" />
    </template>
    <template v-else-if="category === 'entertainment'">
      <path d="M4 5h16v5a2 2 0 0 0 0 4v5H4v-5a2 2 0 0 0 0-4zM12 5v3m0 3v2m0 3v3" />
    </template>
    <template v-else-if="category === 'health'">
      <path d="M9 3h6v6h6v6h-6v6H9v-6H3V9h6z" />
    </template>
    <template v-else-if="category === 'shopping'">
      <path d="M4 8h16l-1 13H5zM9 9V6a3 3 0 0 1 6 0v3" />
    </template>
    <template v-else>
      <circle cx="12" cy="12" r="9" />
      <circle cx="7.5" cy="12" r=".7" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r=".7" fill="currentColor" stroke="none" />
      <circle cx="16.5" cy="12" r=".7" fill="currentColor" stroke="none" />
    </template>
  </svg>
</template>
