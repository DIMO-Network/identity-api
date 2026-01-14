FROM golang:1.25 AS build

RUN useradd -u 10001 dimo

WORKDIR /build
COPY . ./

RUN make build

FROM gcr.io/distroless/base AS final

LABEL maintainer="DIMO <hello@dimo.zone>"

USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /build/bin/identity-api /
COPY --from=build --chown=nonroot:nonroot /build/migrations /migrations

EXPOSE 8080
EXPOSE 8888

ENTRYPOINT ["/identity-api"]

