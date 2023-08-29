FROM debian:bookworm-slim
# Install build dependencies
RUN apt-get update && apt-get install -y binutils squashfs-tools ruby-full make golang rpm
# golang = go 1.19.8
# Install fpm for package building
RUN gem install fpm

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN make build
RUN make build-all
RUN make package-all

ENTRYPOINT ["bash"]
