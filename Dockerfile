FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fethcher ./main.go

RUN useradd -r -u 10001 -g nogroup fethuser


FROM scratch

COPY --from=builder /app/fethcher /fethcher

USER 10001

ENTRYPOINT ["/fethcher"]