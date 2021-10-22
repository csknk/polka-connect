BIN := $(notdir $(CURDIR))
PI_BIN := pi-$(BIN)
SOURCE_FILES := $(wildcard *.go)

${BIN}: ${SOURCE_FILES}
	@echo ${GOBIN}
	@go build -o $@

${PI_BIN}: ${SOURCE_FILES} deps
	@env GOOS=linux GOARCH=arm GOARM=7 go build -o $@

.PHONY: all run clean deps

all: ${BIN} 

build: ${BIN}

pi-build: ${PI_BIN}

run: ${BIN}
	./${BIN}

install: ${BIN}
	@cp ${BIN} ${GOBIN}

clean:
	@go clean
	@rm ${BIN}
