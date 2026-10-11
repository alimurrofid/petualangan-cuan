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
    { name: "Em_Tax", emoji: "🧾" },
    { name: "Em_Shield", emoji: "🛡️" },
    { name: "Em_Gem", emoji: "💎" },
  ],
  Lifestyle: [
    { name: "Em_Pizza", emoji: "🍕" },
    { name: "Em_Burger", emoji: "🍔" },
    { name: "Em_Coffee", emoji: "☕" },
    { name: "Em_Cigarette", emoji: "🚬" },
    { name: "Em_Cart", emoji: "🛒" },
    { name: "Em_Game", emoji: "🎮" },
    { name: "Em_Airplane", emoji: "✈️" },
    { name: "Em_Gift", emoji: "🎁" },
    { name: "Em_Shirt", emoji: "👕" },
    { name: "Em_Movie", emoji: "🎬" },
    { name: "Em_Beer", emoji: "🍺" },
    { name: "Em_Camera", emoji: "📷" },
  ],
  Rumah: [
    { name: "Em_House", emoji: "🏠" },
    { name: "Em_Bulb", emoji: "💡" },
    { name: "Em_Bath", emoji: "🛁" },
    { name: "Em_Bed", emoji: "🛏️" },
    { name: "Em_Tool", emoji: "🛠️" },
    { name: "Em_Trash", emoji: "🗑️" },
    { name: "Em_Plant", emoji: "🪴" },
  ],
    Transportasi: [
    { name: "Em_Toll", emoji: "🛣️" },
    { name: "Em_Parking", emoji: "🅿️" },
    { name: "Em_Car", emoji: "🚗" },
    { name: "Em_Bus", emoji: "🚌" },
    { name: "Em_Motor", emoji: "🏍️" },
    { name: "Em_Train", emoji: "🚆" },
    { name: "Em_Gas", emoji: "⛽" },
  ],
  Kesehatan: [
    { name: "Em_Med", emoji: "💊" },
    { name: "Em_Hosp", emoji: "🏥" },
    { name: "Em_Gym", emoji: "🏋️" },
    { name: "Em_Tooth", emoji: "🦷" },
    { name: "Em_Apple", emoji: "🍎" },
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
  { name: "Cigarette", icon: (LucideIcons as any).Cigarette, label: "Rokok" },
  { name: "IceCream", icon: (LucideIcons as any).IceCream, label: "Dessert" },
  { name: "Wine", icon: (LucideIcons as any).Wine, label: "Minuman" },
  
  // Transportasi
  { name: "Car", icon: (LucideIcons as any).Car, label: "Mobil" },
  { name: "Bike", icon: (LucideIcons as any).Bike, label: "Motor" },
  { name: "Bus", icon: (LucideIcons as any).Bus, label: "Bus/Umum" },
  { name: "Fuel", icon: (LucideIcons as any).Fuel, label: "Bensin" },
  { name: "TrainFront", icon: (LucideIcons as any).TrainFront, label: "Kereta Api" },
  { name: "Package", icon: (LucideIcons as any).Package, label: "Paket / Ongkir" },

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
  { name: "Tv", icon: (LucideIcons as any).Tv, label: "TV Kabel / Streaming" },
  { name: "Palette", icon: (LucideIcons as any).Palette, label: "Seni / Hobi" },

  // Kesehatan & Edukasi
  { name: "Stethoscope", icon: (LucideIcons as any).Stethoscope, label: "Kesehatan" },
  { name: "Dumbbell", icon: (LucideIcons as any).Dumbbell, label: "Olahraga" },
  { name: "GraduationCap", icon: (LucideIcons as any).GraduationCap, label: "Pendidikan" },
  { name: "BookOpen", icon: (LucideIcons as any).BookOpen, label: "Buku" },
  { name: "Pill", icon: (LucideIcons as any).Pill, label: "Obat-obatan" },
  { name: "HeartHandshake", icon: (LucideIcons as any).HeartHandshake, label: "Bantuan Sosial" },

  // Pemasukan & Keuangan
  { name: "Briefcase", icon: (LucideIcons as any).Briefcase, label: "Gaji" },
  { name: "TrendingUp", icon: (LucideIcons as any).TrendingUp, label: "Investasi" },
  { name: "Gift", icon: (LucideIcons as any).Gift, label: "Hadiah" },
  { name: "BadgeDollarSign", icon: (LucideIcons as any).BadgeDollarSign, label: "Bonus" },
  { name: "Heart", icon: (LucideIcons as any).Heart, label: "Donasi" },

    // Pajak & Administrasi
  { name: "Scale", icon: (LucideIcons as any).Scale, label: "Pajak" },
  { name: "FileText", icon: (LucideIcons as any).FileText, label: "Administrasi" },
  { name: "Scale", icon: (LucideIcons as any).Scale, label: "Legal" },

  // Asuransi & Keamanan
  { name: "ShieldCheck", icon: (LucideIcons as any).ShieldCheck, label: "Asuransi" },
  { name: "Lock", icon: (LucideIcons as any).Lock, label: "Keamanan" },

  // Hunian & Sewa
  { name: "Building", icon: (LucideIcons as any).Building, label: "Sewa / Kost" },
  { name: "Warehouse", icon: (LucideIcons as any).Warehouse, label: "Gudang" },

  // Transportasi Tambahan
  { name: "ParkingSquare", icon: (LucideIcons as any).ParkingSquare, label: "Parkir" },
  { name: "CarFront", icon: (LucideIcons as any).CarFront, label: "Tol" },
  { name: "Wrench", icon: (LucideIcons as any).Wrench, label: "Service Kendaraan" },

  // Rumah Tangga
  { name: "WashingMachine", icon: (LucideIcons as any).WashingMachine, label: "Laundry" },
  { name: "Hammer", icon: (LucideIcons as any).Hammer, label: "Perbaikan Rumah" },
  { name: "Trash2", icon: (LucideIcons as any).Trash2, label: "Kebersihan" },
  { name: "Cpu", icon: (LucideIcons as any).Cpu, label: "Software / Hardware" },

  // Langganan & Digital
  { name: "Repeat", icon: (LucideIcons as any).Repeat, label: "Langganan" },
  { name: "Cloud", icon: (LucideIcons as any).Cloud, label: "SaaS / Cloud" },

  // Keluarga & Sosial
  { name: "Users", icon: (LucideIcons as any).Users, label: "Keluarga" },
  { name: "Baby", icon: (LucideIcons as any).Baby, label: "Anak" },
  { name: "PawPrint", icon: (LucideIcons as any).PawPrint, label: "Hewan Peliharaan" },

  // Event & Hiburan Lanjutan
  { name: "Ticket", icon: (LucideIcons as any).Ticket, label: "Tiket / Event" },
  { name: "Hotel", icon: (LucideIcons as any).Hotel, label: "Hotel" },
  { name: "Camera", icon: (LucideIcons as any).Camera, label: "Fotografi" },

  // Keuangan Lanjutan
  { name: "Send", icon: (LucideIcons as any).Send, label: "Transfer" },
  { name: "Target", icon: (LucideIcons as any).Target, label: "Target Tabungan" },
  { name: "BanknoteArrowDown", icon: (LucideIcons as any).BanknoteArrowDown, label: "Utang" },
  { name: "BanknoteArrowUp", icon: (LucideIcons as any).BanknoteArrowUp, label: "Piutang" },
  { name: "CircleFadingArrowUp", icon: (LucideIcons as any).CircleFadingArrowUp, label: "Bayar Utang" },
  { name: "HandCoins", icon: (LucideIcons as any).HandCoins, label: "Terima Piutang" },
  { name: "RefreshCcw", icon: (LucideIcons as any).RefreshCcw, label: "Refund" },
  { name: "AlertTriangle", icon: (LucideIcons as any).AlertTriangle, label: "Denda" },
  { name: "HelpCircle", icon: (LucideIcons as any).HelpCircle, label: "Lain-lain" },
];

// Daftar Icon untuk Tipe Dompet
export const walletIcons: IconItem[] = [
  { name: "Wallet", icon: (LucideIcons as any).Wallet, label: "Dompet Umum" },
  { name: "Banknote", icon: (LucideIcons as any).Banknote, label: "Uang Tunai" },
  { name: "Landmark", icon: (LucideIcons as any).Landmark, label: "Bank" },
  { name: "CreditCard", icon: (LucideIcons as any).CreditCard, label: "Kartu Kredit/Debit" },
  { name: "SmartphoneNfc", icon: (LucideIcons as any).SmartphoneNfc, label: "E-Wallet" },
  { name: "PiggyBank", icon: (LucideIcons as any).PiggyBank, label: "Tabungan" },
  { name: "Vault", icon: (LucideIcons as any).Vault, label: "Brankas" },
  { name: "Bitcoin", icon: (LucideIcons as any).Bitcoin, label: "Kripto" },
  { name: "Briefcase", icon: (LucideIcons as any).Briefcase, label: "Rekening Bisnis" },
  { name: "Coins", icon: (LucideIcons as any).Coins, label: "Koin / Receh" },
  { name: "TrendingUp", icon: (LucideIcons as any).TrendingUp, label: "Investasi" },
  { name: "Globe", icon: (LucideIcons as any).Globe, label: "Rekening Valas" },
  { name: "Shield", icon: (LucideIcons as any).Shield, label: "Dana Darurat" },
  { name: "Archive", icon: (LucideIcons as any).Archive, label: "Dana Cadangan" },
  { name: "Building2", icon: (LucideIcons as any).Building2, label: "Rekening Perusahaan" },
  { name: "HardDrive", icon: (LucideIcons as any).HardDrive, label: "Aset Digital" },
];

export interface BankIconItem {
  name: string;
  label: string;
  logo: string;
  gradient: string;
  textColor: string;
}

export const bankIcons: BankIconItem[] = [
  { name: "Bank_BCA", label: "BCA", logo: "/banks/bca.svg", gradient: "from-[#003893] to-[#0055d4]", textColor: "text-white" },
  { name: "Bank_Mandiri", label: "Mandiri", logo: "/banks/mandiri.svg", gradient: "from-[#003D79] to-[#0a5cb0]", textColor: "text-white" },
  { name: "Bank_BRI", label: "BRI", logo: "/banks/bri.svg", gradient: "from-[#00529C] to-[#0070d6]", textColor: "text-white" },
  { name: "Bank_BNI", label: "BNI", logo: "/banks/bni.svg", gradient: "from-[#005E6A] to-[#008b9c]", textColor: "text-white" },
  { name: "Bank_BTN", label: "BTN", logo: "/banks/btn.svg", gradient: "from-[#012f6b] to-[#015bfd]", textColor: "text-white" },
  { name: "Bank_BSI", label: "BSI", logo: "/banks/bsi.svg", gradient: "from-[#00A39D] to-[#00cfc8]", textColor: "text-white" },
  { name: "Bank_CIMB", label: "CIMB Niaga", logo: "/banks/cimb.svg", gradient: "from-[#7e0000] to-[#ED1C24]", textColor: "text-white" },
  { name: "Bank_Permata", label: "Permata Bank", logo: "/banks/permata.svg", gradient: "from-[#005f59] to-[#009688]", textColor: "text-white" },
  { name: "Bank_Danamon", label: "Bank Danamon", logo: "/banks/danamon.svg", gradient: "from-[#0b3c66] to-[#0066b2]", textColor: "text-white" },
  { name: "Bank_OCBC", label: "Bank OCBC", logo: "/banks/ocbc.svg", gradient: "from-[#991b1b] to-[#dc2626]", textColor: "text-white" },
  { name: "Bank_Mega", label: "Bank Mega", logo: "/banks/mega.svg", gradient: "from-[#b45309] to-[#ea580c]", textColor: "text-white" },
  { name: "Bank_Sinarmas", label: "Bank Sinarmas", logo: "/banks/sinarmas.svg", gradient: "from-[#881337] to-[#e11d48]", textColor: "text-white" },
  { name: "Bank_Muamalat", label: "Bank Muamalat", logo: "/banks/muamalat.svg", gradient: "from-[#3b1259] to-[#6b21a8]", textColor: "text-white" },
  { name: "Bank_Jago", label: "Bank Jago", logo: "/banks/jago.svg", gradient: "from-[#b45309] via-[#d97706] to-[#f59e0b]", textColor: "text-white" },
  { name: "Bank_Seabank", label: "Seabank", logo: "/banks/seabank.svg", gradient: "from-[#d43817] to-[#EE4D2D]", textColor: "text-white" },
  { name: "Bank_Blu", label: "Blu (BCA Digital)", logo: "/banks/blu.svg", gradient: "from-[#0070c9] to-[#0091FF]", textColor: "text-white" },
  { name: "Bank_Jenius", label: "Jenius (BTPN)", logo: "/banks/jenius.svg", gradient: "from-[#007b88] to-[#00A5B5]", textColor: "text-white" },
  { name: "Bank_Allo", label: "Allo Bank", logo: "/banks/allo.svg", gradient: "from-[#0369a1] to-[#0284c7]", textColor: "text-white" },
  { name: "Bank_Line", label: "Line Bank", logo: "/banks/linebank.svg", gradient: "from-[#047857] to-[#10b981]", textColor: "text-white" },
  { name: "Bank_Superbank", label: "Superbank", logo: "/banks/superbank.svg", gradient: "from-[#1e293b] to-[#334155]", textColor: "text-white" },
  { name: "Ew_GoPay", label: "GoPay", logo: "/banks/gopay.svg", gradient: "from-[#0089a8] to-[#00AED6]", textColor: "text-white" },
  { name: "Ew_OVO", label: "OVO", logo: "/banks/ovo.svg", gradient: "from-[#351f6d] to-[#4C3494]", textColor: "text-white" },
  { name: "Ew_DANA", label: "DANA", logo: "/banks/dana.svg", gradient: "from-[#0d6db3] to-[#118EEA]", textColor: "text-white" },
  { name: "Ew_ShopeePay", label: "ShopeePay", logo: "/banks/shopeepay.svg", gradient: "from-[#d43817] to-[#EE4D2D]", textColor: "text-white" },
  { name: "Ew_LinkAja", label: "LinkAja", logo: "/banks/linkaja.svg", gradient: "from-[#ab1218] to-[#EE2E24]", textColor: "text-white" },
  { name: "QRIS", label: "QRIS", logo: "/banks/qris.svg", gradient: "from-[#171717] to-[#262626]", textColor: "text-white" },
];

export const getBankIcon = (name: string | undefined | null): BankIconItem | null => {
  if (!name) return null;
  return bankIcons.find((b) => b.name === name) || null;
};

export const getBankLogo = (name: string | undefined | null): string | undefined => {
  if (!name) return undefined;
  const found = bankIcons.find((b) => b.name === name);
  return found ? found.logo : undefined;
};