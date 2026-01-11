import { GraphQLClient } from 'graphql-request';

const endpoint = import.meta.env.VITE_GRAPHQL_ENDPOINT || 'http://localhost:8080/query';

export const graphqlClient = new GraphQLClient(endpoint, {
  headers: {
    // Add authorization header when needed
  },
});
