export interface Products {
  id: number
  sku: string
  name: string
  price: number
  stock: number
  is_active: boolean
  last_movement_at: string | null
}

export interface ProductsListResponse {
  data: Products[]
  page: number
  limit: number
  total: number
}