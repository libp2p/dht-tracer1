.PHONY: tracer/run
tracer/run:
	@LIBP2P_ALLOW_WEAK_RSA_KEYS=1 go run ./tracedht/tracedht.go --debug --serve :7000

.PHONY: tracer/run-with-logs
tracer/run-with-logs:
	@LIBP2P_ALLOW_WEAK_RSA_KEYS=1 go run ./tracedht/tracedht.go --debug --serve :7000 2>&1 | tee logs.txt
