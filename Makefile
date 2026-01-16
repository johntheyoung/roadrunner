.PHONY: fmt fmt-check test vet lint

fmt:
	gofmt -w .

fmt-check:
	@fmt_out=$$(gofmt -l .); \
	if [ -n "$$fmt_out" ]; then \
		echo "gofmt needed on:"; \
		echo "$$fmt_out"; \
		exit 1; \
	fi

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
