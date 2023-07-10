# identity-api

Generate [SQLBoiler](https://github.com/volatiletech/sqlboiler) models:

```
go run ./cmd/identity-api migrate
make models
```

## Start App

`go run ./cmd/identity-api`

## Create migration

`goose -dir migrations create <migration_name> sql`
