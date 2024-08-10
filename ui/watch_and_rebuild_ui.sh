#!/usr/bin/env bash

# OS: macOS
# watch changes at ui/src and run `npm run build && cp -r dist ../static` when changes are detected

fswatch -o ./src | xargs -n1 -I{} npm run build-and-update
