FROM  golang:1.26-apline as builder

WORKDIR /app 

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o go_auth./cmd/go_auth_cli/main.go



FROM apline:latest 
WORKDIR /root/

COPY --from=builder /app/go_auth . 
CMD ["./go_auth"]

