FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Copy all the source files
COPY . .

# Compile program
RUN CGO_ENABLED=0 GOOS=linux go build -o /ereader-rss

EXPOSE 3000

CMD ["/ereader-rss"]
