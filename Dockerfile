FROM golang:1.16

# Install build dependencies
RUN apt-get update && apt-get install -y binutils squashfs-tools ruby-full make golang rpm

# Install fpm for package building
RUN gem install fpm

WORKDIR /app
COPY ./ ./
RUN go mod download

RUN make build-all
RUN make package-all

ENTRYPOINT ["bash"]
