# Build application
FROM golang:1.24.5 AS build

WORKDIR /app

# Copy contents src directory
COPY src/* ./
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/pod-terminator-controller

# Now copy it into our base image
# Smaller, distroless
FROM gcr.io/distroless/static-debian12
COPY --from=build /build/pod-terminator-controller /app/

# Expose https port
EXPOSE 80

# Run
CMD ["/app/pod-terminator-controller"]