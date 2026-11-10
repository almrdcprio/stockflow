import { createTheme, type MantineColorsTuple } from '@mantine/core'

const brand: MantineColorsTuple = [
  '#eff6ff',
  '#dbeafe',
  '#bfdbfe',
  '#93c5fd',
  '#60a5fa',
  '#3b82f6',
  '#2563eb',
  '#1d4ed8',
  '#1e40af',
  '#1e3a8a',
]

export const theme = createTheme({
  primaryColor: 'brand',

  colors: {
    brand,
  },

  fontFamily: 'Inter, Arial, sans-serif',

  headings: {
    fontFamily: 'Inter, Arial, sans-serif',
    fontWeight: '600',
  },

  defaultRadius: 'md',
})