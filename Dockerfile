FROM golang:1-alpine as builder
WORKDIR /build
COPY . .

# Make sure tests pass before building :)
RUN CGO_ENABLED=0 go test ./...
RUN CGO_ENABLED=0 go build -a -o render-drawio ./cmd/render-drawio

FROM rlespinasse/drawio-desktop-headless
# Enables README support etc. in Github Packages. See: https://docs.github.com/en/packages/guides/about-github-container-registry
LABEL org.opencontainers.image.source=https://github.com/racklet/render-drawio-action

COPY --from=builder /build/render-drawio /
ENTRYPOINT ["/render-drawio"]
