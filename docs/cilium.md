## cilium的kube-proxy代替方案
- Cilium  使用第四种方法（cgroup/connect4）
- Cilium  map分为两部分，cilium_lb4_services_v2：对应k8s的svc资源（带端口），cilium_lb4_backends_v3：对应k8s的endpoint部分（带端口）
- 先查询cilium_lb4_services_v2获得endpoints的数量（count）和key前缀，然后使用key前缀+id（id大于0小于count）查询endpoints，最后将地址和端口改成实际的endpoints的ip和端口。


## Cilium  svc地址转化endpoint地址的方式
举例如果一个svc包含50个endpoints，那么每个endpoint都有两个id（eid和ID），eid范围是1-50，eid帮助负载均衡，ID是endpoint的全局唯一ID。
1. 从cgroup/connect4获取svc地址和port。
2. 使用svc和port向lb4_key查询svc记录，这次查询key是：IP+port+0。
3. 得到value，从value获取endpoint的数量，根据负载均衡算法获取 endpoint的 eid。然后通过IP+port+eid获取endpoint的ID
4. 最后根据endpoint的ID获取endpoint的地址。