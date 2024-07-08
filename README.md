# eproxy
eProxy is a lightweight and efficient replacement for kube-proxy in Kubernetes environments, leveraging eBPF (Extended Berkeley Packet Filter) technology for enhanced performance and flexibility.

eProxy is not a replacement for kube-proxy. As part of a hybrid service proxy, eProxy alleviates the workload of kube-proxy.eProxy can also be part of the solution for macvlan networks.

# How to deploy 

```shell
kubectl apply -f eproxy.yaml
```

# How to use

kubectl apply -f svc.yaml
```shell
apiVersion: v1
kind: Service
metadata:
  labels:
    service.kubernetes.io/service-proxy-name: eproxy
  name: nginx-bpf
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: ClusterIP
```

```shell
kubectl exec -it <pod> bash
curl nginx-bpf:8080
```

# How to build

```shell
docker build -it --rm -v ${eproxy_home}:/root/eproxy registry.cn-hangzhou.aliyuncs.com/secrity/eproxy_build:0.0.1 bash
cd /root/eproxy
make clean all
```