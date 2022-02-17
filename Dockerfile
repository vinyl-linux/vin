FROM golang:1.17

WORKDIR /app
ADD . .

RUN make install
