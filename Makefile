.PHONY: test docs-view

test: gofmt
	for cmd in cmd/*; do \
		(go run $$cmd/*.go -h) || exit 1; \
	done
	make gofmt # to fixup the generated files
	@echo done sanity check CLIs
	go test ./...

docs-view:
	@if ! command -v godoc >/dev/null 2>&1; then \
		echo "Installing godoc..."; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "Starting godoc server on http://localhost:6060"
	@echo "Visit: http://localhost:6060/pkg/github.com/choonkeat/2c2p"
	@godoc -http=:6060

gofmt:
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	goimports -w .
	gofmt -w .

lint:
	golangci-lint run --enable=unused --fix

verify-local: gofmt
	make -f Makefile.local