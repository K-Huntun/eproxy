
all: eproxy ebpfs
	@echo "Build finished."

ebpfs:
	@echo "Build ebpf."
	make -C ebpf/src clean all

eproxy:
	@echo "Build eproxy."
	GOOS=linux GOARCH=amd64 go build -o bin/eproxy github.com/eproxy/cmd/eproxy

clean:
	rm -rf bin/eproxy
	rm -rf ebpf/output/cpnnect.o