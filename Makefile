BINARY=dynamotk

all: build install
build:
	go build -o ${BINARY}
	@echo "build is done."
install:
	@mv ${BINARY} $(GOPATH)/bin
	@echo "${BINARY} is installed successfully."
test:
	go test ./calc/... ./toolkit/... -v