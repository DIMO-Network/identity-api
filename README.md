# identity-api

Run `make` to see some helpful sub-commands:

```
% make

Specify a subcommand:

  run                  Run the app.
  migrate              Run unapplied database migrations.
  sql                  Create a new SQL migration file. Use the NAME variable to set the name: "make sql NAME=dcn_table".
  boil                 Regenerate SQLBoiler models.
  gql                  Regenerate gqlgen code.
  lint                 Run golangci-lint linters.
  tools                Install the correct version of tools used by this project (golangci-lint, goose, sqlboiler, gqlgen).
```

```mermaid
flowchart TD
    Vehicle <--> Aftermarket
    Vehicle <--> Synthetic
    Privilege --> Vehicle
    Vehicle <--> DCN
```

- Vehicles
  - Selection
    - To which vehicles do I have access? Either because I own them or because they are shared with me.
  - Which (non-expired) privileges have been granted on these?
- Aftermarket devices
  - Selection
    - Which devices do I own?
  - Is it paired? To which vehicle?

## Migrations

Add a migrations:
`make tools-goose` Installs the correct version of goose.
`make sql NAME=dcn_table` creates a new migration file.

## Generate GraphQL

`make tools-gqlgen` Installs the correct version of gqlgen.
`make gql` Regenerates the gqlgen code.

## Generate SQLBoiler

`docker-compose up -d db` Starts a Postgres database.
`make tools-boil` Installs the correct version of sqlboiler.
`make migrate` Runs unapplied database migrations.
`make boil` Regenerates the SQLBoiler models.

## License

[Apache 2.0](LICENSE)
