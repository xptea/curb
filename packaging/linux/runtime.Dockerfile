FROM ubuntu:22.04
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    python3 libgtk-3-0 libepoxy0 fonts-dejavu-core ca-certificates ibus-gtk3 \
    build-essential nasm pkg-config libmp3lame-dev libgnutls28-dev zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*
