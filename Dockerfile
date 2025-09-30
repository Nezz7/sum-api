FROM golang:1.25.1-trixie AS build-stage

WORKDIR /app

COPY go.mod /app/go.mod
COPY main.go /app/main.go 

EXPOSE 8080

RUN  go mod download

RUN go build -o /app/sum-api main.go

FROM gcr.io/distroless/base-debian12:nonroot

COPY --from=build-stage /app/sum-api /sum-api 
CMD  ./sum-api

ENTRYPOINT ["/sum-api"]