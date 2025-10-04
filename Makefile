
test:
	go test -race -v ./...

test_octogen:
	cd ./testdata/octogen_tests && go run _run/main.go

autogen_test:
	cd ./testdata/octogen_tests && go run _gen/main.go
