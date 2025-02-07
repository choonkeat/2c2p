.PHONY: test docs-view

test: gofmt
	for cli in cli/*; do \
		(go run $$cli/*.go -h) || exit 1; \
	done
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
	gofmt -w .