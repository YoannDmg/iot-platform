import { ApolloClient, InMemoryCache, HttpLink } from "@apollo/client/core"

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080"

const httpLink = new HttpLink({
  uri: `${API_URL}/query`,
  fetch: (uri, options) => {
    const token = localStorage.getItem("auth_token")
    const headers = new Headers(options?.headers)
    if (token) {
      headers.set("Authorization", `Bearer ${token}`)
    }
    return fetch(uri, { ...options, headers }).then(async (response) => {
      if (response.ok) {
        const clonedResponse = response.clone()
        try {
          const json = await clonedResponse.json()
          if (json.errors) {
            for (const err of json.errors) {
              if (err.extensions?.code === "UNAUTHENTICATED") {
                localStorage.removeItem("auth_token")
                window.location.href = "/login"
                break
              }
            }
          }
        } catch {
          // Response is not JSON, ignore
        }
      }
      return response
    })
  },
})

export const apolloClient = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          devices: {
            keyArgs: ["type", "status"],
            merge(_existing, incoming) {
              return incoming
            },
          },
        },
      },
    },
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: "cache-and-network",
    },
  },
})
