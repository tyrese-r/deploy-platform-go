FROM golang:1.26 AS build


WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY *.go ./

# Run the actual build command here
RUN CGO_ENABLED=0 GOOS=linux go build -o /deploy-platform-go-hello-world

# Copy the binary to empty container
# https://docs.docker.com/build/building/multi-stage/

FROM scratch
COPY --from=build /deploy-platform-go-hello-world /deploy-platform-go-hello-world

ENV PORT=8081
EXPOSE 8081

CMD ["/deploy-platform-go-hello-world"]