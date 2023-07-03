NAME        ?= NatsChat.exe
OUTPUT_BIN  ?= .\bin\${NAME}
SOURCE      ?= .\src
GO_FLAGS    ?=
GO_TAGS     ?= walk_use_cgo
CGO_ENABLED ?= 1
VERSION     ?= $(shell git rev-list --count HEAD)-$(shell git rev-parse --short=7 HEAD)
LD_FLAGS    ?= "-w -s -H=windowsgui -X main.version=${VERSION}"

default: move build

move:
	@echo "Moving output..."
	@IF EXIST ${OUTPUT_BIN} ( @move /y ${OUTPUT_BIN} ${OUTPUT_BIN}.bak )

build:
	@echo "Building from source..."
	@set CGO_ENABLED=${CGO_ENABLED}
	go build -v -tags ${GO_TAGS} ${GO_FLAGS} -ldflags=${LD_FLAGS} -o ${OUTPUT_BIN} ${SOURCE}
