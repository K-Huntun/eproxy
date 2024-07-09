FROM registry.cn-hangzhou.aliyuncs.com/secrity/centos:8
COPY bin/eproxy /eproxy/
COPY ebpf/output/service.o ebpf/