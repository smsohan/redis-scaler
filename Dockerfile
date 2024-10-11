FROM alpine

ARG BINARY

COPY ./build/${BINARY} ./app

ENTRYPOINT ./app