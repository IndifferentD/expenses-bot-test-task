import { createRouter, createWebHistory } from "vue-router";
import ExpensesPage from "../pages/expenses/ExpensesPage.vue";
import UsersPage from "../pages/users/UsersPage.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/expenses" },
    { path: "/expenses", name: "expenses", component: ExpensesPage },
    { path: "/users", name: "users", component: UsersPage },
  ],
});
