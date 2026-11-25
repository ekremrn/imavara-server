FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -p 2 -trimpath -o /out/ ./cmd/...
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
ARG SERVICE=api
COPY --from=build /out/${SERVICE} /usr/local/bin/service
COPY --from=build /out/replay /usr/local/bin/replay
COPY --from=build /out/snapshot /usr/local/bin/snapshot
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/service"]
