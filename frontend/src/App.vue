<script setup lang="ts">
import { ref } from 'vue';
import { Activity, Moon, Sun, CheckCircle2 } from '@lucide/vue';

const isDark = ref(false)

function toggleTheme() {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}
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

        <button
          @click="toggleTheme"
          class="inline-flex items-center space-x-2 px-3 py-1.5 text-xs font-medium rounded-md border
          border-slate-300 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
        >
          <Sun v-if="isDark" :size="16" class="text-amber-400" />
          <Moon v-else :size="16" class="text-slate-600"/>
          <span>{{ isDark ? 'Светлая тема' : 'Темная тема' }}</span>
        </button>
      </header>

      <section class="p-5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm space-y-4">
        <div class="flex items-center space-x-2 text-emerald-600 dark:text-emerald-400">
          <CheckCircle2 :size="20"/>
          <span class="text-sm font-semibold">Фронтенд каркас успешно инициализирован</span>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          Vue 3 (Composition API) + TypeScript 7.0 + Tailwind CSS + Lucide Icons соединены в рабочую среду разработки.
        </p>

        <div class="flex flex-wrap gap-2 pt-2">
          <span class="px-2.5 py-1 text-xs font-mono font-medium rounded bg-indigo-50 dark:bg-indigo-950\60
          text-indigo-700 dark:text-indigo-700 border border-indigo-200 dark:border-indigo-800">
          Vue 3.5+
          </span>

          <span class="px-2.5 py-1 text-xs font-mono font-medium rounded bg-blue-50 dark:bg-blue-950/60
          text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-800">
          TypeScript 7.0
          </span>

          <span class="px-2.5 py-1 text-xs font-mono font-medium rounded bg-cyan-50 dark:bg-cyan-950/60 text-cyan-700 dark:text-cyan-300 border border-cyan-200 dark:border-cyan-800">
            Tailwind CSS
          </span>

          <span class="px-2.5 py-1 text-xs font-mono font-medium rounded bg-violet-50 dark:bg-violet-950/60 text-violet-700 dark:text-violet-300 border border-violet-200 dark:border-violet-800">
            Vite 6+
          </span>
        </div>
      </section>
    </div>
  </main>
</template>
