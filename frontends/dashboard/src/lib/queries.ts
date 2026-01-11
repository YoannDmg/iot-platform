import { gql } from 'graphql-request';

export const GET_DEVICES = gql`
  query GetDevices($page: Int, $pageSize: Int) {
    devices(page: $page, pageSize: $pageSize) {
      devices {
        id
        name
        type
        status
        lastSeen
      }
      total
      page
      pageSize
    }
  }
`;

export const GET_DEVICE = gql`
  query GetDevice($id: ID!) {
    device(id: $id) {
      id
      name
      type
      status
      metadata {
        key
        value
      }
      createdAt
      lastSeen
    }
  }
`;
