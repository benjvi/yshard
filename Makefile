
.PHONY: build test-e2e

build:
	/bin/bash -c "GOOS=darwin GOARCH=amd64 go build -o yshard-darwin-amd64" && \
	/bin/bash -c "GOOS=darwin GOARCH=arm64 go build -o yshard-darwin-arm64" && \
	/bin/bash -c "GOOS=linux GOARCH=amd64 go build -o yshard-linux-amd64" && \
	/bin/bash -c "GOOS=linux GOARCH=arm64 go build -o yshard-linux-arm64" && \
	/bin/bash -c "GOOS=windows GOARCH=amd64 go build -o yshard-windows-amd64"

# assume we always test on linux amd64 for now
test-e2e:
	make build && \
	mkdir -p /tmp/yshard && \
	/bin/bash -c 'cat example.yml | ./yshard-linux-amd64 -g ".group" -o /tmp/yshard' && \
	echo "Are there: 2 docs in a.yml and b.yml, 1 doc in __ungrouped__.yml? If so, test was successful"
