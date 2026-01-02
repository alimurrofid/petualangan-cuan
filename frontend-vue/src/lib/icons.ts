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
    { name: "Em_Chart", emoji: "📈" },
  ],
  Lifestyle: [
    { name: "Em_Pizza", emoji: "🍕" },
    { name: "Em_Burger", emoji: "🍔" },
    { name: "Em_Cart", emoji: "🛒" },
    { name: "Em_Coffee", emoji: "☕" },
    { name: "Em_Game", emoji: "🎮" },
    { name: "Em_Airplane", emoji: "✈️" },
    { name: "Em_Gift", emoji: "🎁" },
    { name: "Em_Shirt", emoji: "👕" },
  ],
  Rumah: [
    { name: "Em_House", emoji: "🏠" },
    { name: "Em_Bulb", emoji: "💡" },
    { name: "Em_Bath", emoji: "🛁" },
    { name: "Em_Bed", emoji: "🛏️" },
    { name: "Em_Tool", emoji: "🛠️" },
  ],
};

export const getEmoji = (name: string | undefined | null): string | null => {
  if (!name) return null;
  
  for (const category of Object.values(emojiCategories)) {
    const found = category.find((e) => e.name === name);
    if (found) return found.emoji;
  }
  
  // Fallback: Jika string adalah emoji native
  if (/\p{Emoji}/u.test(name)) return name;
  
  return null;
};

export { LucideIcons };

export const getIconComponent = (name: string | undefined | null, defaultName?: string): Component | null => {
  if (!name) {
    if (defaultName) return (LucideIcons as any)[defaultName] || null;
    return null;
  }
  return (LucideIcons as any)[name] || (defaultName ? (LucideIcons as any)[defaultName] : null);
};

export interface IconItem {
  name: string;
  icon: Component;
  label: string;
}

export const categoryIcons: IconItem[] = [
  // Makanan & Minuman
  { name: "Utensils", icon: (LucideIcons as any).Utensils, label: "Makan" },
  { name: "Coffee", icon: (LucideIcons as any).Coffee, label: "Kopi/Cafe" },
  { name: "Pizza", icon: (LucideIcons as any).Pizza, label: "Jajan" },
  
  // Transportasi
  { name: "Car", icon: (LucideIcons as any).Car, label: "Mobil" },
  { name: "Bike", icon: (LucideIcons as any).Bike, label: "Motor" },
  { name: "Bus", icon: (LucideIcons as any).Bus, label: "Bus/Umum" },
  { name: "Fuel", icon: (LucideIcons as any).Fuel, label: "Bensin" },

  // Belanja
  { name: "ShoppingCart", icon: (LucideIcons as any).ShoppingCart, label: "Belanja Harian" },
  { name: "ShoppingBag", icon: (LucideIcons as any).ShoppingBag, label: "Shopping" },
  { name: "Store", icon: (LucideIcons as any).Store, label: "Toko" },

  // Tagihan & Rumah
  { name: "Home", icon: (LucideIcons as any).Home, label: "Rumah" },
  { name: "Zap", icon: (LucideIcons as any).Zap, label: "Listrik" },
  { name: "Droplet", icon: (LucideIcons as any).Droplet, label: "Air" },
  { name: "Wifi", icon: (LucideIcons as any).Wifi, label: "Internet" },
  { name: "Phone", icon: (LucideIcons as any).Phone, label: "Pulsa" },
  
  // Hiburan & Hobi
  { name: "Gamepad2", icon: (LucideIcons as any).Gamepad2, label: "Game" },
  { name: "Film", icon: (LucideIcons as any).Film, label: "Nonton" },
  { name: "Music", icon: (LucideIcons as any).Music, label: "Musik" },
  { name: "Plane", icon: (LucideIcons as any).Plane, label: "Traveling" },

  // Kesehatan & Edukasi
  { name: "Stethoscope", icon: (LucideIcons as any).Stethoscope, label: "Kesehatan" },
  { name: "Dumbbell", icon: (LucideIcons as any).Dumbbell, label: "Olahraga" },
  { name: "GraduationCap", icon: (LucideIcons as any).GraduationCap, label: "Pendidikan" },
  { name: "BookOpen", icon: (LucideIcons as any).BookOpen, label: "Buku" },

  // Pemasukan & Keuangan
  { name: "Briefcase", icon: (LucideIcons as any).Briefcase, label: "Gaji" },
  { name: "TrendingUp", icon: (LucideIcons as any).TrendingUp, label: "Investasi" },
  { name: "Gift", icon: (LucideIcons as any).Gift, label: "Hadiah" },
  { name: "BadgeDollarSign", icon: (LucideIcons as any).BadgeDollarSign, label: "Bonus" },
  { name: "Heart", icon: (LucideIcons as any).Heart, label: "Donasi" },
];

// Daftar Icon untuk Tipe Dompet
export const walletIcons: IconItem[] = [
  { name: "Wallet", icon: (LucideIcons as any).Wallet, label: "Dompet Umum" },
  { name: "Banknote", icon: (LucideIcons as any).Banknote, label: "Uang Tunai" },
  { name: "Landmark", icon: (LucideIcons as any).Landmark, label: "Bank" },
  { name: "CreditCard", icon: (LucideIcons as any).CreditCard, label: "Kartu Kredit/Debit" },
  { name: "SmartphoneNfc", icon: (LucideIcons as any).SmartphoneNfc, label: "E-Wallet" }, // GoPay, OVO, dll
  { name: "PiggyBank", icon: (LucideIcons as any).PiggyBank, label: "Tabungan" },
  { name: "Vault", icon: (LucideIcons as any).Vault, label: "Brankas" },
  { name: "Bitcoin", icon: (LucideIcons as any).Bitcoin, label: "Kripto" },
];