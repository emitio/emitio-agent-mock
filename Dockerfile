FROM ubuntu:xenial-20180412

COPY ./target/emitio-agent-mock_linux_amd64 /usr/local/bin/emitio-agent-mock

ENTRYPOINT [ "emitio-agent-mock" ]