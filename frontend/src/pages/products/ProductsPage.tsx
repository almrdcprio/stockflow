import { useEffect, useState } from 'react'
import { useForm } from '@mantine/form'
import {
  Alert,
  Badge,
  Button,
  Card,
  Group,
  Loader,
  Modal,
  NumberInput,
  Pagination,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
} from '@mantine/core'
import { AlertCircle, Plus, Search, Pencil } from 'lucide-react'

import { apiGet, apiPost } from '../../services/api'
import { formatCurrency } from '../../utils/format'
import type { Products, ProductsListResponse } from '../../types/products'

interface ProductFormValues {
  sku: string
  name: string
  price: number | string
  stock: number | string
}

function ProductsPage() {
  const [products, setProducts] = useState<ProductsListResponse | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [createOpened, setCreateOpened] = useState(false)
  const [saving, setSaving] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [refresh, setRefresh] = useState(0)

  const limit = 10

  const form = useForm<ProductFormValues>({
    mode: 'controlled',
    initialValues: {
      sku: '',
      name: '',
      price: 0,
      stock: 0,
    },
    validate: {
      sku: (value) =>
        value.trim().length === 0 ? 'SKU is required' : null,
      name: (value) =>
        value.trim().length === 0 ? 'Product name is required' : null,
      price: (value) =>
        Number.isInteger(Number(value)) && Number(value) >= 0
          ? null
          : 'Price must be a non-negative integer',
      stock: (value) =>
        Number.isInteger(Number(value)) && Number(value) >= 0
          ? null
          : 'Stock must be a non-negative integer',
    },
  })

  useEffect(() => {
    let active = true

    const loadProducts = async () => {
      try {
        setLoading(true)
        setError(null)

        const params = new URLSearchParams({
          search,
          page: String(page),
          limit: String(limit),
        })

        const data = await apiGet<ProductsListResponse>(
          `/api/products?${params.toString()}`,
        )

        if (active) {
          setProducts(data)
        }
      } catch (err) {
        if (active) {
          setError(
            err instanceof Error
              ? err.message
              : 'Failed to load products',
          )
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void loadProducts()

    return () => {
      active = false
    }
  }, [search, page, refresh])

  const handleCreateProduct = async (values: ProductFormValues) => {
    try {
      setSaving(true)
      setCreateError(null)

      await apiPost<Products>('/api/products', {
        sku: values.sku.trim(),
        name: values.name.trim(),
        price: Number(values.price),
        stock: Number(values.stock),
      })

      form.reset()
      setCreateOpened(false)
      setPage(1)
      setRefresh((current) => current + 1)
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : 'Failed to create product',
      )
    } finally {
      setSaving(false)
    }
  }

  const totalPages = products
    ? Math.ceil(products.total / limit)
    : 0

  return (
    <Stack gap="lg">
      <Group justify="space-between" align="flex-end">
        <div>
          <Title order={2}>Products</Title>
          <Text c="dimmed" size="sm" mt={4}>
            Manage and monitor your inventory products.
          </Text>
        </div>

        <Button
          leftSection={<Plus size={16} />}
          onClick={() => {
            setCreateError(null)
            setCreateOpened(true)
          }}
        >
          Add Product
        </Button>
      </Group>

      <Card withBorder radius="md" padding="lg">
        <Stack gap="md">
          <Group justify="space-between">
            <Text fw={600}>Product List</Text>

            <Text size="sm" c="dimmed">
              {products?.total ?? 0} products
            </Text>
          </Group>

          <TextInput
            placeholder="Search by SKU or product name..."
            leftSection={<Search size={16} />}
            value={search}
            onChange={(event) => {
              setSearch(event.currentTarget.value)
              setPage(1)
            }}
          />

          {loading && (
            <Group justify="center" py="xl">
              <Loader size="sm" />
              <Text size="sm" c="dimmed">
                Loading products...
              </Text>
            </Group>
          )}

          {!loading && error && (
            <Alert
              color="red"
              icon={<AlertCircle size={18} />}
              title="Failed to load products"
            >
              {error}
            </Alert>
          )}

          {!loading && !error && products && (
            <>
              {products.data.length === 0 ? (
                <Text c="dimmed" ta="center" py="xl">
                  No products found.
                </Text>
              ) : (
                <Table.ScrollContainer minWidth={700}>
                  <Table verticalSpacing="md" highlightOnHover>
                    <Table.Thead>
                      <Table.Tr>
                        <Table.Th>SKU</Table.Th>
                        <Table.Th>Product Name</Table.Th>
                        <Table.Th>Price</Table.Th>
                        <Table.Th>Stock</Table.Th>
                        <Table.Th>Status</Table.Th>
                      </Table.Tr>
                    </Table.Thead>

                    <Table.Tbody>
                      {products.data.map((product) => (
                        <Table.Tr key={product.id}>
                          <Table.Td>
                            <Text fw={500} size="sm">
                              {product.sku}
                            </Text>
                          </Table.Td>

                          <Table.Td>{product.name}</Table.Td>
                          <Table.Td>{formatCurrency(product.price)}</Table.Td>
                          <Table.Td>{product.stock}</Table.Td>

                          <Table.Td>
                            <Badge
                              color={product.is_active ? 'green' : 'gray'}
                              variant="light"
                            >
                              {product.is_active ? 'Active' : 'Inactive'}
                            </Badge>
                          </Table.Td>
                        </Table.Tr>
                      ))}
                    </Table.Tbody>
                  </Table>
                </Table.ScrollContainer>
              )}

              {totalPages > 1 && (
                <Group justify="flex-end">
                  <Pagination
                    value={page}
                    onChange={setPage}
                    total={totalPages}
                  />
                </Group>
              )}
            </>
          )}
        </Stack>
      </Card>

      <Modal
        opened={createOpened}
        onClose={() => {
          setCreateOpened(false)
          setCreateError(null)
          form.reset()
        }}
        title="Add New Product"
        centered
      >
        <form onSubmit={form.onSubmit(handleCreateProduct)}>
          <Stack gap="md">
            {createError && (
              <Alert color="red" icon={<AlertCircle size={18} />}>
                {createError}
              </Alert>
            )}

            <TextInput
              label="SKU"
              placeholder="e.g. SKU-013"
              required
              {...form.getInputProps('sku')}
            />

            <TextInput
              label="Product Name"
              placeholder="Enter product name"
              required
              {...form.getInputProps('name')}
            />

            <NumberInput
              label="Price (IDR)"
              placeholder="Enter product price"
              min={0}
              allowDecimal={false}
              required
              {...form.getInputProps('price')}
            />

            <NumberInput
              label="Initial Stock"
              placeholder="Enter initial stock"
              min={0}
              allowDecimal={false}
              required
              {...form.getInputProps('stock')}
            />

            <Group justify="flex-end" mt="sm">
              <Button
                variant="default"
                onClick={() => {
                  setCreateOpened(false)
                  setCreateError(null)
                  form.reset()
                }}
              >
                Cancel
              </Button>

              <Button type="submit" loading={saving}>
                Create Product
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Stack>
  )
}

export default ProductsPage