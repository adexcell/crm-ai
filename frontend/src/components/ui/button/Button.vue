<script setup lang="ts">
import { computed } from 'vue';
import { Loader2 } from '@lucide/vue';
import { cn } from '@/lib/utils';

export interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
  class?: string
}

const props = withDefaults(defineProps<ButtonProps>(), {
  variant: 'primary',
  size: 'md',
  loading: false,
  disabled: false,
  type: 'button',
})

const baseStyles = 'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none active:scale-[0.98]'
const variantStyles: Record<string, string> = {
  primary: 'bg-brand-500 hover:bg-brand-600 text-white shadow-sm focus:ring-brand-500',
  secondary: 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 focus:ring-slate-400',
  outline: 'border border-slate-300 dark:border-slate-700 bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 focus:ring-slate-400',
  ghost: 'bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 focus:ring-slate-400',
  danger: 'bg-rose-600 hover:bg-rose-700 text-white shadow-sm focus:ring-rose-500',
}

const sizeStyles: Record<string, string> = {
  sm: 'h-8 px-2.5 text-xs gap-1.5',
  md: 'h-9 px-4 text-sm gap-2',
  lg: 'h-11 px-6 text-base gap-2.5',
}

const buttonClasses = computed(() => {
  return cn(
    baseStyles,
    variantStyles[props.variant],
    sizeStyles[props.size],
    props.class,
  )
})

const iconSizes: Record<string, number> = {
  sm: 14,
  md: 16,
  lg: 18
}

</script>

<template>
  <button
    v-bind:type="type"
    v-bind:disabled="disabled || loading"
    v-bind:class="buttonClasses"
  >
    <Loader2
      v-if="loading"
      class="animate-spin shrink-0"
      v-bind:size="iconSizes[size]"
    />
    <slot />
  </button>
</template>
