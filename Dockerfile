FROM golang:1.27.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -o /server ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -o /richmenu ./cmd/richmenu

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /server /server
COPY --from=build /richmenu /richmenu
COPY assets/richmenu/ /assets/richmenu/
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
