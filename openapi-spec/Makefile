.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: serve
serve:
## serve: Serve the sites locally on port 8080 using Docker.
	docker run \
		--rm \
		-p 8080:8080 \
		-v ${PWD}:/srv \
		caddy:latest \
		caddy file-server --listen=:8080 --root=/srv

.PHONY: mock
mock:
## mock: Run a mock server based on the OpenAPI definition
	docker run \
		--rm \
		--init \
		-p 4010:4010 \
		-v ${PWD}:/srv \
		stoplight/prism:4 \
		mock -h 0.0.0.0 "/srv/openapi.yaml"

.PHONY: lint
lint:
## lint: Lint the OpenAPI spec with Spectral
	npx @stoplight/spectral-cli lint ${CURDIR}/openapi.yaml
	npx prettier --check .

.PHONY: format
format:
## format: Formats both Markdown documents and YAML documents to preferred repository style.
	npx prettier --write .
	yq -i 'sort_keys(.components.schemas)' openapi.yaml
	yq -i 'sort_keys(.components.responses)' openapi.yaml
	yq -i 'sort_keys(.components.parameters)' openapi.yaml
	yq -i 'sort_keys(.paths)' openapi.yaml

