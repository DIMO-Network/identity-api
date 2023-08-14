.PHONY: help run migrate boil gql sql

NAME?="new"

run: ## Run the app.
	go run ./cmd/identity-api
migrate: ## Run database migrations.
	go run ./cmd/identity-api migrate
sql: ## Create a new SQL migration file. Use the NAME variable to set the name: "make sql NAME=dcn_table".
	goose -dir migrations create $(NAME) sql
boil: ## Generate SQLBoiler models.
	sqlboiler psql --no-tests --wipe
gql: ## Generate gqlgen code.
	go run github.com/99designs/gqlgen generate

help:
	@echo "\nSpecify a subcommand:\n"
	@grep -hE '^[0-9a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[m %s\n", $$1, $$2}'
	@echo ""

.DEFAULT_GOAL := help
