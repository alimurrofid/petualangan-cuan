import * as LucideIcons from "lucide-vue-next";
import { type Component } from "vue";

export interface EmojiItem {
  name: string;
  emoji: string;
}

export type EmojiCategoryMap = Record<string, EmojiItem[]>;

export const emojiCategories: EmojiCategoryMap = {
  Keuangan: [
    { name: "Em_MoneyBag", emoji: "💰" },
    { name: "Em_DollarBill", emoji: "💵" },
    { name: "Em_Card", emoji: "💳" },
    { name: "Em_Bank", emoji: "🏦" },
    { name: "Em_MoneyWing", emoji: "💸" },
    { name: "Em_Coin", emoji: "🪙" },
  ],
  Lifestyle: [
    { name: "Em_Pizza", emoji: "🍕" },
    { name: "Em_Cart", emoji: "🛒" },
    { name: "Em_Coffee", emoji: "☕" },
    { name: "Em_Game", emoji: "🎮" },
    { name: "Em_Airplane", emoji: "✈️" },
    { name: "Em_Gift", emoji: "🎁" },
  ],
  Simbol: [
    { name: "Em_Star", emoji: "⭐" },
    { name: "Em_Fire", emoji: "🔥" },
    { name: "Em_Lock", emoji: "🔒" },
    { name: "Em_Check", emoji: "✅" },
    { name: "Em_Idea", emoji: "💡" },
  ],
};

export const getEmoji = (name: string | undefined | null): string | null => {
  if (!name) return null;
  
  for (const category of Object.values(emojiCategories)) {
    const found = category.find((e) => e.name === name);
    if (found) return found.emoji;
  }
  
  if (/\p{Emoji}/u.test(name)) return name;
  
  return null;
};

export const getIconComponent = (name: string | undefined | null, defaultName?: string): Component | null => {
  if (!name) {
    if (defaultName) return (LucideIcons as any)[defaultName] || null;
    return null;
  }
  return (LucideIcons as any)[name] || (defaultName ? (LucideIcons as any)[defaultName] : null);
};

export { LucideIcons };
