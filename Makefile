build:
	go build -o kafka

run-server:
	go run . server

run-client:
	go run . client

clean:
	rm -f kafka
