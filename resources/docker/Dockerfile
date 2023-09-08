FROM golang:1.21 as build

WORKDIR /src/identity-api

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o /bin/identity-api .

FROM gcr.io/distroless/static-debian11 AS release

COPY --from=build /bin/identity-api /identity-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/identity-api"]
