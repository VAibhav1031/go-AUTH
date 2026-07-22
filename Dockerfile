FROM  golang:1.26-alpine as builder

WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o go_auth ./cmd/go_auth_cli/


FROM alpine:latest 
WORKDIR /app

COPY --from=builder /app/go_auth . 
CMD ["./go_auth"]

