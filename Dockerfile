FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

# Create directory for SSL certificates
RUN mkdir -p /etc/mycelium/certs

# Add the SSL certificates
COPY certs/cert.pem /etc/mycelium/certs/cert.pem
COPY certs/key.pem /etc/mycelium/certs/key.pem

# Set appropriate permissions for the certificates
RUN chmod 600 /etc/mycelium/certs/key.pem

ADD mycelium /usr/bin/mycelium

CMD ["mycelium"]