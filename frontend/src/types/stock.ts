export interface StockMovement {
  id: number
  product_id: number
  type: 'in' | 'out'
  qty: number
  note: string | null
  created_at: string
}

export interface StockMovementsResponse {
  data: StockMovement[]
}