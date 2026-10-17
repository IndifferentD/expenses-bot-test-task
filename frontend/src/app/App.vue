<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router';
</script>

<template>
  <div class="app-shell">
    <header class="site-header">
      <div class="site-header__inner">
        <RouterLink class="brand" to="/expenses" aria-label="Семейные расходы — главная">
          <span class="brand__mark" aria-hidden="true">
            <svg viewBox="0 0 28 28" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M5 9.5h18v13H5zM5 9.5V7a2 2 0 0 1 2-2h13M18 16h5M18 16h-2" />
            </svg>
          </span>
          <span>Семейные расходы</span>
        </RouterLink>

        <nav class="site-nav" aria-label="Основная навигация">
          <RouterLink to="/expenses" active-class="site-nav__link--active" class="site-nav__link">Траты</RouterLink>
          <RouterLink to="/users" active-class="site-nav__link--active" class="site-nav__link">Семья</RouterLink>
        </nav>
      </div>
    </header>

    <RouterView />
  </div>
</template>
