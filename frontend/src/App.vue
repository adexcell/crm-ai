<script setup lang="ts">
import { ref, onMounted } from 'vue';
import {
  Activity,
  Moon,
  Sun,
  CheckCircle2,
  XCircle,
  RefreshCw,
  Database,
  Layers,
} from '@lucide/vue';

import { checkReadiness } from './api/modules/health';
import type { HealthStatus } from './api/types';

import { useDark } from '@vueuse/core';

const isDark = useDark();
function toggleTheme() {
  isDark.value = !isDark.value;
};
const isLoading = ref(false);
const healthData = ref<HealthStatus | null>(null);
const errorMessage = ref<string | null>(null);

async function fetchHealth() {
  isLoading.value = true
  errorMessage.value = null

  try {
    const data = await checkReadiness()
    healthData.value = data
  } catch (err: any) {
    errorMessage.value = err.Message || "Не удалось связаться с бэкендом"
    healthData.value = null
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchHealth()
})
</script>

<template>
  <main class="min-h-screen p-8 transition-colors duration-200">
    <div class="max-w-4xl mx-auto space-y-6">

      <header class="flex items-center justify-between pb-6 border-b border-slate-200 dark:border-slate-800">

        <div class="flex items-center space-x-3">

          <div class="p-2.5 bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white rounded-lg shadow-sm">
            <Activity :size="24" />
          </div>

          <div>
            <h1 class="text-x1 font-bold tracking-tight">CRM AI Platform</h1>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Спринт 0: Инфраструктурный фундамент и фронтенд
            </p>
          </div>

        </div>

        <div class="flex items-center space-x-3">

          <button
            @click="fetchHealth"
            :disabled="isLoading"
            class="inline-flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium rounded-md bg-brand-500
            hover:bg-brand-600 disabled:opacity-50 disabled:cursor-not-allowed text-white shadow-sm transition"
          >
            <RefreshCw :size="14" :class="{ 'animate-spin': isLoading }" />
            <span>{{ isLoading ? 'Проверяем...' : 'Обновить статус' }}</span>

          </button>

          <button
            @click="toggleTheme"
            class="inline-flex items-center space-x-2 px-3 py-1.5 text-xs font-medium rounded-md border
            border-slate-300 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            <Sun v-if="isDark" :size="16" class="text-amber-400" />
            <Moon v-else :size="16" class="text-slate-600"/>
            <span>{{ isDark ? 'Светлая тема' : 'Темная тема' }}</span>
          </button>

        </div>

      </header>

      <section
        v-if="errorMessage"
        class="p-5 rounded-x1 border border-rose-200 dark:border-rose-900 bg-rose-50 dark:bg-rose-950/40 text-rose-900
        dark:text-rose-200 space-y-3"
      >
        <div class="flex items-center space-x-2 font-semibold text-sm text-rose-700 dark:text-rose-300">
          <XCircle :size="18" />
          <span>Бэкенд пока не отвечает</span>
        </div>

        <p class="text-xs leading-relaxed">
          {{ errorMessage }}
        </p>

        <p class="text-xs text-rose-600 dark:text-rose-400 font-mono bf-white/60 dark:bg-black/30 p-2.5 rounded border
        border-rose-200 dark:border-rose-800">
          Подсказка: запустите FastAPI в соседней вкладке терминала:<br>
          cd backend-python && uv run uvicorn src.main:app --reload --port 8000
        </p>
      </section>

      <section v-else-if="healthData" class="space-y-4">

        <div class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex items-center
        justify-between">

          <div class="flex items-center space-x-3">

            <div class="p-2 bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400 rounded-lg">
              <CheckCircle2 :size="20" />
            </div>

            <div>
              <h2 class="text-sm font-semibold">API Сервер не активен</h2>
              <p class="text-xs text-slate-500 dark:text-slate-400">Эндпоинт /api/v1/health/ready ответил статусом 200 OK</p>
            </div>
          </div>

          <span class="px-2.5 py-1 text-xs font-mono font-medium rounded-full bg-emerald-100 dark:bg-emerald-900/60
          text-emerald-800 dark:text-emerald-300">
            ready: ok
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

          <!-- Postgres -->
          <div class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm
          space-y-3">

            <div class="flex items-center justify-between">

              <div class="flex items-center space-x-2 text-slate-800 dark:text-slate-200">
                <Database :size="18" class="text-indigo-500" />
                <span class="text-sm font-semibold">PostgreSQL 16 + pgvector</span>
              </div>

              <span class="inline-flex items-center space-x-1.5 px-2 py-0.5 text-xs font-medium rounded bg-emerald-50
              dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                <span>{{ healthData.postgres }}</span>
              </span>

            </div>

            <p class="text-xs text-slace-500 dark:text-slate-400">
              Реляционные данные, пул asyncpg и векторный поиск HNSW
            </p>

          </div>

          <!-- Redis -->
          <div class="p-4 rounded-x1 border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm space-y-3">

            <div class="flex items-center justify-between">

              <div class="flex items-center space-x-2 text-slate-800 dark:text-slate-200">
                <Layers :size="18" class="text-rose-500" />
                <span class="text-sm font-semibold">Redis 7 (In-Memory)</span>
              </div>

              <span class="inline-flex items-center space-x-1.5 px-2 py-0.5 text-xs font-medium rounded bg-emerald-50
              dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                <span>{{ healthData.redis }}</span>
              </span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Кэширование сессий, Pub/Sub шина и WebSocket rate limiting
            </p>
          </div>

        </div>
      </section>

    </div>
  </main>
</template>
