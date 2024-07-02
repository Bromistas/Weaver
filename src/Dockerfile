# Start from the official Golang image
FROM golang:1.21-alpine

ENV ADDRESS="127.0.0.1:50051"
ENV PORT="50051"
ENV WAIT_TIME="3"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.work go.work.sum ./
COPY common/go.mod common/go.sum ./common/
COPY router/go.mod router/go.sum ./router/
COPY node/go.mod node/go.sum ./node/
COPY protos/go.mod protos/go.sum ./protos/
COPY scrapper/go.mod scrapper/go.sum ./scrapper/
COPY rpc_node/go.mod ./rpc_node/
COPY storage/go.mod ./storage/

# Download all the dependencies
RUN go mod download

# Copy the entire workspace into the container
COPY . .

# Set the working directory to the storage project
WORKDIR /app/storage

# Build the Go app
RUN go build -o /storage

EXPOSE 50051

CMD ["/storage"]

## Start from the official Golang image
#FROM golang:1.21-alpine
#
#ENV ADDRESS="127.0.0.1:50051"
#ENV PORT="50051"
#ENV WAIT_TIME="3"
#
## Set the Current Working Directory inside the container
#WORKDIR /app
#
#RUN mkdir "storage"
#
## Set the working directory to the storage project
#WORKDIR /app/storage
#
#COPY storage/go.mod .
#
## Download all the dependencies
#RUN go mod download
#
#
#WORKDIR /app
#
## Copy the entire workspace into the container
#COPY . .
#
#WORKDIR /app/storage
#
#
## Build the Go app
#RUN go build -o /storage
#
#EXPOSE 50051
#
#CMD ["/storage"]