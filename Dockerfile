# Sử dụng image Golang
FROM golang:1.20

# Set thư mục làm việc
WORKDIR /app

# Copy tất cả file từ thư mục hiện tại
COPY . .

# Lấy các dependency
RUN go mod tidy

# Build ứng dụng
RUN go build -o main .

# Expose cổng chạy server
EXPOSE 8080

# Chạy ứng dụng
CMD ["./main"]