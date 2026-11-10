
import { useEffect, useState } from 'react'
import {
    Alert,
    Button,
    Card,
    Loader,
    Select,
    Stack,
    Text,
    Title,
    Badge,
    Table,
    NumberInput,
    SegmentedControl,
    Textarea,
} from '@mantine/core'
import { AlertCircle } from 'lucide-react'
import { apiGet, apiPost } from '../../services/api'
import type { Products, ProductsListResponse } from '../../types/products'
import type { StockMovement, StockMovementsResponse } from '../../types/stock'

function StocksPage() {
    const [products, setProducts] = useState<Products[]>([])
    const [selectedProduct, setSelectedProduct] = useState<string | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [refresh, setRefresh] = useState(0)

    const [movements, setMovements] = useState<StockMovement[]>([])
    const [movementLoading, setMovementLoading] = useState(false)
    const [movementError, setMovementError] = useState<string | null>(null)
    const [movementRefresh, setMovementRefresh] = useState(0)

    const [adjustmentType, setAdjustmentType] = useState<'in' | 'out'>('in')
    const [quantity, setQuantity] = useState<number | string>(1)
    const [note, setNote] = useState('')
    const [saving, setSaving] = useState(false)
    const [adjustmentError, setAdjustmentError] = useState<string | null>(null)
    const [adjustmentSuccess, setAdjustmentSuccess] = useState(false)

    useEffect(() => {
        let active = true

        async function fetchProducts() {
        try {
            setLoading(true)
            setError(null)

            const result = await apiGet<ProductsListResponse>(
            '/api/products?status=all&page=1&limit=100',
            )

            if (active) {
            setProducts(result.data)
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

        fetchProducts()

        return () => {
        active = false
        }
    }, [refresh])

    useEffect(() => {
        if (!selectedProduct) {
            setMovements([])
            setMovementError(null)
            return
        }

        let active = true

        async function fetchMovements() {
            try {
                setMovementLoading(true)
                setMovementError(null)

                const result = await apiGet<StockMovementsResponse>(
                    `/api/products/${selectedProduct}/movements`,
                )

                if (active) {
                    setMovements(result.data)
                }
            } catch (err) {
            if (active) {
                setMovementError(
                err instanceof Error
                    ? err.message
                    : 'Failed to load movements',
                )
            }
            } finally {
                if (active) {
                    setMovementLoading(false)
                }
            }
        }

        fetchMovements()

        return () => {
            active = false
        }
    }, [selectedProduct, movementRefresh])

    const productOptions = products.map((product) => ({
        value: String(product.id),
        label: `${product.sku} - ${product.name}`,
    }))

    async function handleStockAdjustment() {
        if (!selectedProduct) return

        const qty = Number(quantity)

        if (!Number.isInteger(qty) || qty <= 0) {
            setAdjustmentError('Quantity must be a positive integer.')
            return
        }

        try {
            setSaving(true)
            setAdjustmentError(null)
            setAdjustmentSuccess(false)

            await apiPost(
            `/api/products/${selectedProduct}/stock`,
            {
                type: adjustmentType,
                qty,
                note: note.trim(),
            },
            )

            setQuantity(1)
            setNote('')
            setAdjustmentSuccess(true)

            // Refresh product stock and movement history
            setRefresh((value) => value + 1)
            setMovementRefresh((value) => value + 1)
        } catch (err) {
            setAdjustmentError(
            err instanceof Error
                ? err.message
                : 'Failed to adjust stock',
            )
        } finally {
            setSaving(false)
        }
    }

    return (
        <Stack gap="lg">
        <div>
            <Title order={2}>Stock Movements</Title>
            <Text c="dimmed">
            Track your stock movements.
            </Text>
        </div>

        <Card withBorder radius="md" padding="lg">
            <Stack gap="md">
            <Title order={4}>Select Product</Title>

            {loading ? (
                <Loader size="sm" />
            ) : error ? (
                <Alert
                icon={<AlertCircle size={16} />}
                color="red"
                title="Failed to load products"
                >
                <Stack gap="xs">
                    <Text size="sm">{error}</Text>
                    <Button
                    size="xs"
                    variant="light"
                    color="red"
                    onClick={() => setRefresh((value) => value + 1)}
                    >
                    Retry
                    </Button>
                </Stack>
                </Alert>
            ) : (
                <Select
                placeholder="Choose a product"
                data={productOptions}
                value={selectedProduct}
                onChange={setSelectedProduct}
                searchable
                clearable
                nothingFoundMessage="No products found"
                />
            )}

            {selectedProduct && (
                <Card withBorder radius="md" padding="lg">
                    <Stack gap="md">
                    <Title order={4}>Movement History</Title>

                    {movementLoading ? (
                        <Loader size="sm" />
                    ) : movementError ? (
                        <Alert
                        color="red"
                        title="Failed to load movements"
                        icon={<AlertCircle size={16} />}
                        >
                        <Stack gap="xs">
                            <Text size="sm">{movementError}</Text>
                            <Button
                            size="xs"
                            variant="light"
                            color="red"
                            onClick={() =>
                                setMovementRefresh((value) => value + 1)
                            }
                            >
                            Retry
                            </Button>
                        </Stack>
                        </Alert>
                    ) : movements.length === 0 ? (
                        <Text c="dimmed" ta="center" py="lg">
                        No stock movements found for this product.
                        </Text>
                    ) : (
                        <Table.ScrollContainer minWidth={650}>
                        <Table highlightOnHover>
                            <Table.Thead>
                            <Table.Tr>
                                <Table.Th>Date</Table.Th>
                                <Table.Th>Type</Table.Th>
                                <Table.Th>Quantity</Table.Th>
                                <Table.Th>Note</Table.Th>
                            </Table.Tr>
                            </Table.Thead>

                            <Table.Tbody>
                            {movements.map((movement) => (
                                <Table.Tr key={movement.id}>
                                <Table.Td>
                                    {new Date(movement.created_at).toLocaleString(
                                    'id-ID',
                                    )}
                                </Table.Td>

                                <Table.Td>
                                    <Badge
                                    color={
                                        movement.type === 'in' ? 'green' : 'red'
                                    }
                                    >
                                    {movement.type.toUpperCase()}
                                    </Badge>
                                </Table.Td>

                                <Table.Td>
                                    <Text
                                    fw={600}
                                    c={movement.type === 'in' ? 'green' : 'red'}
                                    >
                                    {movement.type === 'in' ? '+' : '-'}
                                    {movement.qty}
                                    </Text>
                                </Table.Td>

                                <Table.Td>
                                    {movement.note || '-'}
                                </Table.Td>
                                </Table.Tr>
                            ))}
                            </Table.Tbody>
                        </Table>
                        </Table.ScrollContainer>
                    )}
                    </Stack>
                </Card>
            )}

            {selectedProduct && (
                <Card withBorder radius="md" padding="lg">
                    <Stack gap="md">
                    <Title order={4}>Stock Adjustment</Title>

                    <SegmentedControl
                        fullWidth
                        value={adjustmentType}
                        onChange={(value) => {
                        setAdjustmentType(value as 'in' | 'out')
                        setAdjustmentError(null)
                        setAdjustmentSuccess(false)
                        }}
                        data={[
                        { label: 'Stock In', value: 'in' },
                        { label: 'Stock Out', value: 'out' },
                        ]}
                    />

                    <NumberInput
                        label="Quantity"
                        placeholder="Enter quantity"
                        min={1}
                        allowDecimal={false}
                        value={quantity}
                        onChange={setQuantity}
                    />

                    <Textarea
                        label="Note"
                        placeholder="Enter adjustment note (optional)"
                        value={note}
                        onChange={(event) => setNote(event.currentTarget.value)}
                        minRows={3}
                    />

                    {adjustmentError && (
                        <Alert color="red" title="Adjustment Failed">
                        {adjustmentError}
                        </Alert>
                    )}

                    {adjustmentSuccess && (
                        <Alert color="green" title="Success">
                        Stock adjustment completed successfully.
                        </Alert>
                    )}

                    <Button
                        fullWidth
                        loading={saving}
                        disabled={!selectedProduct}
                        onClick={handleStockAdjustment}
                    >
                        Submit Adjustment
                    </Button>
                    </Stack>
                </Card>
            )}

            </Stack>
        </Card>
        </Stack>
    )
}

export default StocksPage
