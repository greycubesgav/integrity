FROM golang:1.23.3

# Tell the container we're noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y binutils squashfs-tools ruby-full make golang rpm sudo asciinema fonts-liberation

# https://github.com/asciinema/agg/releases/download/v1.4.3/agg-aarch64-unknown-linux-gnu
# Add asciinema

# Install fpm for package building
RUN gem install fpm

# Add agg to render asciinema casts to gifs
RUN curl -L -o /bin/agg https://github.com/asciinema/agg/releases/download/v1.4.3/agg-aarch64-unknown-linux-gnu
RUN chmod +x /bin/agg

# Setup a non-root user
ENV USER=go
RUN useradd --create-home --home-dir "/home/${USER}/" --shell /bin/bash --uid 1000 --user-group -G users "${USER}"
RUN echo "${USER} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/${USER} && chmod 440 "/etc/sudoers.d/${USER}"
USER ${USER}

# Add the application
WORKDIR /app
COPY --chown=${USER}:${USER} ./ ./
RUN go mod download

RUN make test
RUN make build-all
RUN make package-all

# Install the latest version of integrity for usage later
RUN sudo dpkg -i /app/pkgs/integrity_*_arm64.deb

ENTRYPOINT ["bash"]
