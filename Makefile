.PHONY: bin
bin: target/emitio-agent-mock_linux_amd64

SRC_FULL = $(shell find . -type f -not -path './target/*' -not -path './.*')
target/emitio-agent-mock_linux_amd64: $(SRC_FULL)
	@GOOS="linux" GOARCH="amd64" go build -o $@ cmd/emitio-agent-mock/main.go

.PHONY: image
image:
	docker build . -t emitio/emitio-agent-mock

.PHONY: push
push:
	docker push emitio/emitio-agent-mock
