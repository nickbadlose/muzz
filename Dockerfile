ARG GOLANG_VERSION
ARG ALPINE_VERSION

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as build

# set the workdir, this has to be outside to GOPATH so modules work correctly.
WORKDIR /project

COPY . .

# build a binary & move to /var
RUN CGO_ENABLED=0 go build -o /var/api ./cmd/main.go

# expose initial ports
EXPOSE 3000

# --- production stage ---
# https://hub.docker.com/_/alpine/
FROM alpine:${ALPINE_VERSION} AS production

RUN apk --update add ca-certificates tzdata

# copy the built binary from the build process.
COPY --from=build /var/api /var/api

# initial healthcheck to perform on the production build.
HEALTHCHECK --interval=5m --timeout=3s \
  CMD curl -f http://localhost:3000/status || exit 1

# set the work directory.
WORKDIR /var

# run the binary.
CMD ./api

# --- development stage ---
# https://hub.docker.com/_/golang/
FROM build AS development

# expose debug api port.
EXPOSE 3000

# either go run main.go
ENTRYPOINT [ "go", "run", "./cmd/main.go" ]