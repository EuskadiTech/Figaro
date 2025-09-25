# Copy SSL certificates for HTTPS requests (if needed)
# Since we're using scratch as base image, we need to get certificates from a temporary alpine image
FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

FROM scratch

# Copy the pre-built binary from the build context
# The binary will be placed in the build context by the CI/CD pipeline
ARG TARGETARCH
COPY figaro-linux-${TARGETARCH} /figaro

# Copy SSL certificates from alpine image
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8080

# Set default environment variables
ENV PORT=8080
ENV HOST=0.0.0.0
ENV DATA_DIR=/data

# Create volume for data persistence
VOLUME ["/data"]

# Run the binary
ENTRYPOINT ["/figaro"]