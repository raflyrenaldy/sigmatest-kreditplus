# Use the official Golang 1.20 image as the base image
FROM golang:1.20

LABEL maintainer="Orgosys Private Limited <shubham@orgosys.com>"
# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and download the dependencies
#COPY go.mod .
#COPY go.sum .
COPY . ./
RUN go mod tidy

# Copy the project files into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Set the environment variables for MongoDB
COPY .env .

# Start the application
CMD ["./main"]