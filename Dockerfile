FROM golang:1.27.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 go build -trimpath -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /server /server
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
