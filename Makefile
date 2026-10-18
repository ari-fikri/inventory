.PHONY: test test-bdd run clean

test:
	go test -v ./internal/...

test-bdd:
	go test -v ./test/...

run:
	go run cmd/server/main.go

clean:
	rm -f server coverage.out *.db
