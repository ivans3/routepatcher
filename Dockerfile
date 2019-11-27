# Step 1 - Build the executable 
FROM golang:alpine AS builder
RUN apk add -U git

# Copy the local package files into the container's workspace
COPY . /routepatcher/
WORKDIR /routepatcher/

# Get dependencies
RUN go get -d -v

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/routepatcher

# Step 2 - build a smaller image 
FROM scratch

COPY --from=builder /go/bin/routepatcher /go/bin/routepatcher

# Run the outyet command by default when the container starts.
ENTRYPOINT [ "/go/bin/routepatcher" ]

