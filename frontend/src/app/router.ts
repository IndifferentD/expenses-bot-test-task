import { createRouter, createWebHistory } from 'vue-router';
import ExpensesPage from '../pages/expenses/ExpensesPage.vue';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/expenses' },
    { path: '/expenses', name: 'expenses', component: ExpensesPage },
  ],
});
