
## Bpf Map 设计

> bpf map 设计遵循 cilium的设计

### service map

```shell
struct lb_service_key {
   __be32 address;       /* Service 地址 */
   __be16 dport;     /* 端口,如果是0,那么表示任意端口都转化 */
   __u16 backend_slot;    /* endpoint迭代,如果是0表示svc本身,用于帮助确定 */
   __u8 proto;       /* 协议 */
   __u8 scope;       /* 不做了解 */
   __u8 pad[2];      /* 冗余 */
};

struct lb4_service {
	union {
		__u32 backend_id;	/* Backend ID in lb4_backends */
		__u32 affinity_timeout;	/* In seconds, only for svc frontend */
		__u32 l7_lb_proxy_port;	/* In host byte order, only when flags2 && SVC_FLAG_L7LOADBALANCER */
	};
	__u16 count;
	__u16 rev_nat_index;	/* Reverse NAT ID in lb4_reverse_nat */
	__u8 flags;
	__u8 flags2;
	__u8  pad[2];
};

```

### endpoints map

## 资源监控逻辑

* 定时刷新
* 缓存设计

### service 更新逻辑

1. k8s informer watch 事件 // 重写inform接口,使用filterWatch过滤service
2. 添加事件
3. 更新事件（无更新事件）
4. 删除事件
   1. 和bpf cache对比事件
   2. 更新bpf内容，更新失败，则不更新bpf cache

### endpoint 更新逻辑

1. k8s informer watch 事件// 重写inform接口,使用filterWatch过滤service
2. 添加事件
   3. count ++,添加新的endpoint
3. 更新事件（无更新事件）
   4. 更新endpoint
4. 删除事件
   1. 将末尾的endpoint替换 删除的endpoint
   2. 删除末尾的endpoint，count--

1. 和bpf cache对比事件
2. 更新bpf内容，更新失败，则不更新bpf cache

## 兼容性范围
> 本产品属于试验性质

### kubernetes版本

### cgroup版本

### 内核版本