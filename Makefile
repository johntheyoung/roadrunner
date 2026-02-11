.PHONY: fmt fmt-check test test-agent-smoke vet lint skill-sync skill-publish

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

test-agent-smoke:
	./scripts/agent-smoke.sh

vet:
	go vet ./...

lint:
	golangci-lint run ./...

skill-sync:
	cp SKILL.md skill/SKILL.md

VERSION_STRIPPED := $(patsubst v%,%,$(VERSION))

skill-publish: skill-sync
	clawdhub publish $(CURDIR)/skill --slug roadrunner --name "Roadrunner" $(if $(VERSION),--version $(VERSION_STRIPPED),)
