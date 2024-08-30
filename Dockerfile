FROM golang:1.20

# Install build dependencies
RUN apt-get update && apt-get install -y binutils squashfs-tools ruby-full make golang rpm sudo

# Install fpm for package building
RUN gem install fpm

WORKDIR /app
COPY ./ ./
RUN go mod download

RUN make test
RUN make build-all
RUN make package-all

ENTRYPOINT ["bash"]
