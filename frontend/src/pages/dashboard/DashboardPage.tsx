import { useEffect, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Group,
  Loader,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
} from '@mantine/core'
import {
  AlertCircle,
  ArrowDownToLine,
  Package,
  Wallet,
  Warehouse,
} from 'lucide-react'

import { apiGet } from '../../services/api'
import type { DashboardSummary } from '../../types/dashboard'

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(value)
}

function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [retry, setRetry] = useState(0)

  useEffect(() => {
    let active = true

    const loadSummary = async () => {
      try {
        setLoading(true)
        setError(null)

        const data = await apiGet<DashboardSummary>(
          '/api/dashboard/summary',
        )

        if (active) {
          setSummary(data)
        }
      } catch (err) {
        if (active) {
          setError(
            err instanceof Error
              ? err.message
              : 'Failed to load dashboard data',
          )
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void loadSummary()

    return () => {
      active = false
    }
  }, [retry])

  const stats = summary
    ? [
        {
          label: 'Total Products',
          value: summary.total_products.toLocaleString('id-ID'),
          icon: Package,
          color: 'blue',
        },
        {
          label: 'Active Products',
          value: summary.active_products.toLocaleString('id-ID'),
          icon: Warehouse,
          color: 'green',
        },
        {
          label: 'Stock Value',
          value: formatCurrency(summary.stock_value),
          icon: Wallet,
          color: 'violet',
        },
        {
          label: 'Low Stock Products',
          value: summary.low_stock_products.toLocaleString('id-ID'),
          icon: ArrowDownToLine,
          color: 'orange',
        },
      ]
    : []

  return (
    <Stack gap="lg">
      <div>
        <Title order={2}>Dashboard Overview</Title>
        <Text c="dimmed" size="sm" mt={4}>
          Monitor your warehouse performance and inventory.
        </Text>
      </div>

      {loading && (
        <Group justify="center" py="xl">
          <Loader />
          <Text c="dimmed">Loading dashboard data...</Text>
        </Group>
      )}

      {!loading && error && (
        <Alert
          icon={<AlertCircle size={18} />}
          title="Failed to load dashboard"
          color="red"
        >
          <Stack gap="sm">
            <Text size="sm">{error}</Text>

            <Button
              variant="light"
              color="red"
              size="xs"
              onClick={() => setRetry((current) => current + 1)}
              w="fit-content"
            >
              Try Again
            </Button>
          </Stack>
        </Alert>
      )}

      {!loading && !error && summary && (
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }}>
          {stats.map((stat) => {
            const Icon = stat.icon

            return (
              <Card
                key={stat.label}
                padding="lg"
                radius="md"
                withBorder
                shadow="xs"
              >
                <Group justify="space-between" align="flex-start">
                  <Stack gap={6}>
                    <Text size="sm" c="dimmed">
                      {stat.label}
                    </Text>

                    <Title order={3}>{stat.value}</Title>
                  </Stack>

                  <ThemeIcon
                    color={stat.color}
                    variant="light"
                    size="xl"
                    radius="md"
                  >
                    <Icon size={20} />
                  </ThemeIcon>
                </Group>
              </Card>
            )
          })}
        </SimpleGrid>
      )}
    </Stack>
  )
}

export default DashboardPage