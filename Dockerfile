
FROM golang:1.25.1 AS build-stage

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sum-api .

# Using Multi-stage build to remove unnecessary dependency
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /sum-api /sum-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/sum-api"]