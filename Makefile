GO ?= go

.PHONY: test bench vet fmt

test:
	$(GO) test ./...

bench:
	$(GO) test -bench=. -benchmem ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -l .
