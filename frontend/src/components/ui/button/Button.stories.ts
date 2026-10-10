import type { Meta, StoryObj } from '@storybook/vue3';
import { Button } from './index';

const meta = {
  title: 'UI Primitives/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline', 'ghost', 'danger'],
      description: 'Вариант цветовой схемы кнопки',
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Размер кнопки (sm для таблиц CRM, md - стандарт, lg - формы)',
    },
    loading: {
      control: 'boolean',
      description: 'Показывает вращающийся спиннер и блокирует нажатие',
    },
    disabled: {
      control: 'boolean',
      description: 'Отключает кнопку',
    },
  },
  args: {
    default: 'Нажми меня',
  },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  args: {
    variant: 'primary',
    default: 'Основное действие',
  },
}

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    default: 'Второстепенное действие',
  },
}

export const Outline: Story = {
  args: {
    variant: 'outline',
    default: 'Контурная кнопка',
  },
}

export const Ghost: Story = {
  args: {
    variant: 'ghost',
    default: 'Прозрачная кнопка',
  },
}

export const Danger: Story = {
  args: {
    variant: 'danger',
    default: 'Удалить сделку',
  },
}

export const Disabled: Story = {
  args: {
    variant: 'secondary',
    disabled: true,
    default: 'Недоступно'
  },
}
