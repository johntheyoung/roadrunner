.PHONY: fmt fmt-check test vet lint skill-sync skill-publish

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

skill-sync:
	cp SKILL.md skill/SKILL.md

skill-publish: skill-sync
	clawdhub publish skill
