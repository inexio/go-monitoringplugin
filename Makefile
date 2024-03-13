TEST_ARGS=

test:
	go test ${TEST_ARGS} ./...

build:
	go build -ldflags="-s -w" ./

clean:
	rm -f check_wg
