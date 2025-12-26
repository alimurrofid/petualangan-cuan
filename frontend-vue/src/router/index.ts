import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import AdminLayout from "@/layouts/AdminLayout.vue";

const routes: RouteRecordRaw[] = [
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
        path: "wallet",
        name: "Wallet",
        component: () => import("@/views/wallet/Index.vue"),
      },
      {
        path: "category",
        name: "Category",
        component: () => import("@/views/category/Index.vue"),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
