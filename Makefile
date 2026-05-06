GO ?= go

.PHONY: test bench vet fmt

test:
	$(GO) test -timeout 120s ./...

bench:
	$(GO) test -timeout 120s -bench=. -benchmem ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -l .
