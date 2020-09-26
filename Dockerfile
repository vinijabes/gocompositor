FROM golang:1.13

RUN \ 
    apt-get update && apt-get upgrade -y && \
    apt-get install -y libgstreamer1.0-0 \
            gstreamer1.0-plugins-base \
            gstreamer1.0-plugins-good \
            gstreamer1.0-plugins-bad \
            gstreamer1.0-plugins-ugly \
            gstreamer1.0-libav \
            gstreamer1.0-doc \
            gstreamer1.0-tools \
            libgstreamer1.0-dev && \
            libgstreamer-plugins-base1.0-dev && \
            libgstreamer-plugins-bad1.0-dev &&\
    apt-get install -y xvfb \
            x11-xserver-utils
            

RUN DOCKERIZE_VERSION=v0.6.1 \
	&& apt-get update \
	&& apt-get install -y wget \
	&& wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
	&& mkdir -p /gocompositor

WORKDIR /gocompositor