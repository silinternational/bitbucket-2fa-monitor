FROM golang:1.21

RUN curl -o- -L https://slss.io/install | VERSION=3.34.0 bash && \
  mv $HOME/.serverless/bin/serverless /usr/local/bin && \
  ln -s /usr/local/bin/serverless /usr/local/bin/sls

RUN mkdir -p /app
WORKDIR /app

COPY ./ .
RUN go get ./...
