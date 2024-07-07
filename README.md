# eproxy
eProxy is a lightweight and efficient replacement for kube-proxy in Kubernetes environments, leveraging eBPF (Extended Berkeley Packet Filter) technology for enhanced performance and flexibility.

# How to deploy 

```go
bpftool cgroup detach /run/eproxy/cgroupv2  connect4 id 10
```


# How to build

```shell
docker build -it --rm -v ${eproxy_home}:/root/eproxy registry.cn-hangzhou.aliyuncs.com/secrity/eproxy_build:0.0.1 bash
cd /root/eproxy
make clean all
```