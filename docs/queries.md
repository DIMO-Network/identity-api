### Show all of my vehicles

```graphql
{
  vehicles(
    filterBy: {privileged: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"},
    first: 10
  ) {
    edges {
      node {
        tokenId
        definition {
          make
          model
          year
        }
        owner
        mintedAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

Here we're asking for 10 cars to which `0xd8…045` has access. The default sort is descending by token id, so this query will produce the 10 most recently minted cars.

For each of these cars, we are asking for the token id, owner, time of mint; and make, model, and year.

The elements `first`, `edges`, and `pageInfo` come from the [Relay cursor spec](https://relay.dev/graphql/connections.htm).
