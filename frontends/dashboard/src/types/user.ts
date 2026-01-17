// Types partagés pour les utilisateurs - Source unique de vérité

export interface User {
  id: string
  email: string
  name: string
  role: string
  createdAt: number
  lastLogin: number | null
  isActive: boolean
}

export interface AuthPayload {
  token: string
  user: User
}

export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  email: string
  password: string
  name: string
  role?: string
}
