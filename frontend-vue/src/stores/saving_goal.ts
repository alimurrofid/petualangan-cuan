import { defineStore } from "pinia";
import { ref } from "vue";
import api from "@/lib/api";
import { useSwal } from "@/composables/useSwal";

export const useSavingGoalStore = defineStore("savingGoal", () => {
    const goals = ref<any[]>([]);
    const isLoading = ref(false);
    const swal = useSwal();

    const fetchGoals = async () => {
        isLoading.value = true;
        try {
            const response = await api.get('/api/saving-goals');
            goals.value = response.data.data;
        } catch (error) {
            console.error("Failed to fetch saving goals", error);
        } finally {
            isLoading.value = false;
        }
    };

    const createGoal = async (payload: FormData | any) => {
        try {
            await api.post('/api/saving-goals', payload);
            await fetchGoals();
            return true;
        } catch (error: any) {
            console.error("Failed to create saving goal", error);
            swal.error("Gagal", error.response?.data?.error || "Gagal membuat target menabung");
            return false;
        }
    };

    const addContribution = async (goalId: number, payload: any) => {
        try {
            await api.post(`/api/saving-goals/${goalId}/contributions`, payload);
            await fetchGoals();
            return true;
        } catch (error: any) {
            console.error("Failed to add contribution", error);
            swal.error("Gagal", error.response?.data?.error || "Gagal menabung");
            throw error;
        }
    };

    return {
        goals,
        isLoading,
        fetchGoals,
        createGoal,
        addContribution,
    };
});
