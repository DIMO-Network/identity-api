.PHONY: run orm gql

run:
        go run ./cmd/identity-api
orm:
	sqlboiler psql --no-tests --wipe
gql:
        go run github.com/99designs/gqlgen generate
