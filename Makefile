build:
	go build -o kafka ./cmd/kafka

run-server:
	go run ./cmd/kafka broker

run-client:
	go run ./cmd/kafka client

clean:
	rm -f kafka
