default: help

.PHONY: generate-ts
generate-ts: # Generate TypeScript package
	@docker run --rm -ti -v $(pwd):/repo registry.com/tools/openapi-generator generate \
	    -i /repo/openapi/gate/app/openapi.yaml \
	    -g typescript \
	    -o /repo/build \
	    --package-name app-api \
	    --additional-properties=npmName=app-api \
	    --additional-properties=npmVersion=0.0.1 \
	    --additional-properties=platform=deno


.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done