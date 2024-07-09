# eproxy
eProxy is a lightweight and efficient replacement for kube-proxy in Kubernetes environments, leveraging eBPF (Extended Berkeley Packet Filter) technology for enhanced performance and flexibility.

eProxy is not a replacement for kube-proxy. As part of a hybrid service proxy, eProxy alleviates the workload of kube-proxy.eProxy can also be part of the solution for macvlan networks.

# How to deploy 

```shell
git clone https://github.com/K-Huntun/eproxy.git
cd eproxy/deploy
kubectl apply -f deploy.yaml
```

# How to use

kubectl create ns nginx
kubectl apply -f svc.yaml
```shell
apiVersion: v1
kind: Service
metadata:
  labels:
    service.kubernetes.io/service-proxy-name: eproxy
  name: nginx-bpf
  namespace: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx-deploy
  namespace: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        ports:
        - containerPort: 80
        name: nginx
```

```shell
kubectl exec -it <other-pod> bash
curl nginx-bpf:8080
```

# 卸载
Due to the fact that eproxy mounts the BTF file internally within the container, 
when eproxy is unmounted or when the container is killed for other reasons, 
both the eBPF program and map data of eproxy will disappear. Upon container restart, they will be reloaded again.

```shell
kubectl delete -f deploy.yaml
```

# How to build

```shell
docker build -it --rm -v ${eproxy_home}:/root/eproxy registry.cn-hangzhou.aliyuncs.com/secrity/eproxy_build:0.0.1 bash
cd /root/eproxy
make clean all
```