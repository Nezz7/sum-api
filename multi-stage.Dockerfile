
FROM golang:1.25.1 AS build-stage

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY main.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /sum-api

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /sum-api /sum-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/sum-api"]