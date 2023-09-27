FROM golang:1.20

RUN apt-get update && apt-get install ruby vim-common -y

ENV LIBPCAP_VERSION=1.10.4

RUN apt-get install flex bison -y
RUN wget http://www.tcpdump.org/release/libpcap-${LIBPCAP_VERSION}.tar.gz  \
    && tar xzf libpcap-${LIBPCAP_VERSION}.tar.gz  \
    && cd libpcap-${LIBPCAP_VERSION}  \
    && ./configure  \
    && make install

WORKDIR /go/src/net-capture/
