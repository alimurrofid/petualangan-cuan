<script setup lang="ts">
import { 
  Card, 
  CardContent, 
  CardDescription, 
  CardHeader, 
  CardTitle 
} from '@/components/ui/card'
import { 
  ArrowUpIcon, 
  ArrowDownIcon, 
  DollarSign, 
  CreditCard,
  Activity
} from 'lucide-vue-next'

const recentTransactions = [
  {
    id: 1,
    name: "Gaji Bulanan",
    email: "Kantor Utama",
    amount: "+Rp 15.000.000",
    status: "income"
  },
  {
    id: 2,
    name: "Belanja Bulanan",
    email: "Supermarket",
    amount: "-Rp 2.500.000",
    status: "expense"
  },
  {
    id: 3,
    name: "Netflix Subscription",
    email: "Hiburan",
    amount: "-Rp 180.000",
    status: "expense"
  },
  {
    id: 4,
    name: "Freelance Project",
    email: "Klien A",
    amount: "+Rp 3.500.000",
    status: "income"
  },
  {
    id: 5,
    name: "Service Motor",
    email: "Transportasi",
    amount: "-Rp 450.000",
    status: "expense"
  },
]
</script>

<template>
  <div class="flex-1 space-y-4">
    <div class="flex items-center justify-between space-y-2">
      <h2 class="text-3xl font-bold tracking-tight">Dashboard</h2>
      <div class="flex items-center space-x-2">
        </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Saldo</CardTitle>
          <DollarSign class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">Rp 45.231.000</div>
          <p class="text-xs text-muted-foreground">+20.1% dari bulan lalu</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pemasukan (Bulan Ini)</CardTitle>
          <ArrowUpIcon class="h-4 w-4 text-emerald-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-emerald-600">+Rp 18.500.000</div>
          <p class="text-xs text-muted-foreground">+4% dari bulan lalu</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pengeluaran (Bulan Ini)</CardTitle>
          <ArrowDownIcon class="h-4 w-4 text-rose-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-rose-600">-Rp 3.130.000</div>
          <p class="text-xs text-muted-foreground">-12% dari bulan lalu</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Transaksi Aktif</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">+573</div>
          <p class="text-xs text-muted-foreground">Sejak jam terakhir</p>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
      
      <Card class="col-span-4">
        <CardHeader>
          <CardTitle>Overview Keuangan</CardTitle>
          <CardDescription>Grafik pendapatan vs pengeluaran tahun ini.</CardDescription>
        </CardHeader>
        <CardContent class="pl-2">
          <div class="h-75 flex items-center justify-center bg-muted/20 rounded-md border border-dashed">
            <span class="text-muted-foreground">Area Grafik (Akan dipasang Chart.js nanti)</span>
          </div>
        </CardContent>
      </Card>

      <Card class="col-span-3">
        <CardHeader>
          <CardTitle>Transaksi Terakhir</CardTitle>
          <CardDescription>5 transaksi terbaru Anda.</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-8">
            <div v-for="item in recentTransactions" :key="item.id" class="flex items-center">
              <div class="flex h-9 w-9 items-center justify-center rounded-full bg-muted border">
                <CreditCard class="h-4 w-4 text-foreground" />
              </div>
              <div class="ml-4 space-y-1">
                <p class="text-sm font-medium leading-none">{{ item.name }}</p>
                <p class="text-sm text-muted-foreground">{{ item.email }}</p>
              </div>
              <div 
                class="ml-auto font-medium" 
                :class="item.status === 'income' ? 'text-emerald-600' : 'text-rose-600'"
              >
                {{ item.amount }}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>