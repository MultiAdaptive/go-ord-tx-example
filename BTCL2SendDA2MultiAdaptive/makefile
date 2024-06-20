BINARY_NAME=sendBTCDA
SOURCE_FILE=main.go

all: build

build:
	go build -o $(BINARY_NAME) $(SOURCE_FILE)

clean:
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run
