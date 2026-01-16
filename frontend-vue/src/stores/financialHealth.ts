import { defineStore } from "pinia";
import axios from "@/lib/api";

export interface FinancialHealthRatio {
  name: string;
  value: number;
  formatted_value: string;
  target: string;
  status: "Sehat" | "Waspada" | "Bahaya";
  description: string;
}

export interface FinancialHealthResponse {
  overall_score: number;
  overall_status: "Sehat" | "Waspada" | "Bahaya";
  ratios: FinancialHealthRatio[];
}

export const useFinancialHealthStore = defineStore("financialHealth", {
  state: () => ({
    data: null as FinancialHealthResponse | null,
    isLoading: false,
  }),
  actions: {
    async fetchFinancialHealth() {
      this.isLoading = true;
      try {
        const response = await axios.get("/api/financial-health");
        this.data = response.data.data;
      } catch (error) {
        console.error("Failed to fetch financial health:", error);
      } finally {
        this.isLoading = false;
      }
    },
  },
});
