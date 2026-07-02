.PHONY: build run clean test reset

BINARY := tubely
DB_PATH ?= ./tubely.db

build:
	go build -o $(BINARY)

run: build
	timeout 5m ./$(BINARY) &

test:
	go test ./...

clean:
	rm -f $(BINARY)

reset:
	sqlite3 $(DB_PATH) "DELETE FROM refresh_tokens; DELETE FROM users; DELETE FROM videos;"
