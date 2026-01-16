import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import AdminLayout from "@/layouts/AdminLayout.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/auth/Login.vue"),
  },
  {
    path: "/register",
    name: "Register",
    component: () => import("@/views/auth/Register.vue"),
  },
  {
    path: "/auth/google/callback",
    name: "GoogleCallback",
    component: () => import("@/views/auth/GoogleCallback.vue"),
  },
  {
    path: "/",
    component: AdminLayout,
    redirect: "/dashboard",
    children: [
      {
        path: "dashboard",
        name: "Dashboard",
        component: () => import("@/views/Dashboard.vue"),
      },
      {
         path: "chat",
         name: "Chat",
         component: () => import("@/views/chat/Index.vue"),
      },
      {
        path: "wallet",
        name: "Wallet",
        component: () => import("@/views/wallet/Index.vue"),
      },
      {
        path: "category",
        name: "Category",
        component: () => import("@/views/category/Index.vue"),
      },
      {
        path: "transaction",
        name: "Transaction",
        component: () => import("@/views/transaction/Index.vue"),
      },
      {
        path: "calendar",
        name: "Calendar",
        component: () => import("@/views/calendar/Index.vue"),
      },
      {
        path: "report",
        name: "Report",
        component: () => import("@/views/report/Index.vue"),
      },
      {
         path: "setting",
         name: "Setting",
         component: () => import("@/views/setting/Index.vue"),
      },
      {
        path: "debt",
        name: "Debt",
        component: () => import("@/views/debt/Index.vue"),
      },
      {
        path: "wishlist",
        name: "Wishlist",
        component: () => import("@/views/wishlist/Index.vue"),
      },
      {
        path: "saving-goal",
        name: "SavingGoal",
        component: () => import("@/views/saving_goal/Index.vue"),
      },
      {
        path: "financial-health",
        name: "FinancialHealth",
        component: () => import("@/views/financial-health/Index.vue"),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, _, next) => {
  const token = localStorage.getItem('token');
  const publicPages = ['/login', '/register', '/auth/google/callback'];
  const authRequired = !publicPages.includes(to.path);

  if (authRequired && !token) {
    next('/login');
  } else if ((to.path === '/login' || to.path === '/register') && token) {
     next('/dashboard');
  } else {
    next();
  }
});

export default router;
