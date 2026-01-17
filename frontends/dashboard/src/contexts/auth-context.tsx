import { createContext, useContext, useCallback, type ReactNode } from "react"
import { useQuery, useMutation, useApolloClient } from "@apollo/client/react"
import {
  GET_ME,
  LOGIN,
  REGISTER,
  type GetMeResponse,
  type LoginResponse,
  type LoginVariables,
  type RegisterResponse,
  type RegisterVariables,
} from "@/graphql"
import type { User, LoginInput, RegisterInput, AuthPayload } from "@/types/user"

const AUTH_TOKEN_KEY = "auth_token"

interface AuthContextValue {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (input: LoginInput) => Promise<AuthPayload | null>
  register: (input: RegisterInput) => Promise<AuthPayload | null>
  logout: () => Promise<void>
  loginLoading: boolean
  registerLoading: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const client = useApolloClient()

  // Current user query
  const { data, loading: userLoading } = useQuery<GetMeResponse>(GET_ME, {
    errorPolicy: "ignore",
  })

  // Login mutation
  const [loginMutation, { loading: loginLoading }] = useMutation<
    LoginResponse,
    LoginVariables
  >(LOGIN)

  // Register mutation
  const [registerMutation, { loading: registerLoading }] = useMutation<
    RegisterResponse,
    RegisterVariables
  >(REGISTER)

  const login = useCallback(
    async (input: LoginInput) => {
      const result = await loginMutation({ variables: { input } })
      if (result.data?.login) {
        localStorage.setItem(AUTH_TOKEN_KEY, result.data.login.token)
        await client.refetchQueries({ include: [GET_ME] })
        return result.data.login
      }
      return null
    },
    [loginMutation, client]
  )

  const register = useCallback(
    async (input: RegisterInput) => {
      const result = await registerMutation({ variables: { input } })
      if (result.data?.register) {
        localStorage.setItem(AUTH_TOKEN_KEY, result.data.register.token)
        await client.refetchQueries({ include: [GET_ME] })
        return result.data.register
      }
      return null
    },
    [registerMutation, client]
  )

  const logout = useCallback(async () => {
    localStorage.removeItem(AUTH_TOKEN_KEY)
    await client.clearStore()
  }, [client])

  const value: AuthContextValue = {
    user: data?.me ?? null,
    isAuthenticated: !!data?.me,
    isLoading: userLoading,
    login,
    register,
    logout,
    loginLoading,
    registerLoading,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
