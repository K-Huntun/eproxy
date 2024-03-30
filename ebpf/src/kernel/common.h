// Supplement vmlinux
#ifndef _LINUX_COMMON_H
#define _LINUX_COMMON_H

struct lb4_key {
	__be32 address;		/* Service virtual IPv4 address */
	__be16 dport;		/* L4 port filter, if unset, all ports apply */
	__u16 backend_slot;	/* Backend iterator, 0 indicates the svc frontend */
	__u8 proto;		/* L4 protocol, currently not used (set to 0) */
	__u8 scope;		/* LB_LOOKUP_SCOPE_* for externalTrafficPolicy=Local */
	__u8 pad[2];
};

struct lb4_backend {
	__be32 address;		/* Service endpoint IPv4 address */
	__be16 port;		/* L4 port filter */
	__u8 proto;		/* L4 protocol, currently not used (set to 0) */
	__u8 flags;
	__u16 cluster_id;	/* With this field, we can distinguish two
				 * backends that have the same IP address,
				 * but belong to the different cluster.
				 */
	__u8 pad[2];
};

struct lb4_service {
	union {
		__u32 backend_id;	/* Backend ID in lb4_backends */
		__u32 affinity_timeout;	/* In seconds, only for svc frontend */
		__u32 l7_lb_proxy_port;	/* In host byte order, only when flags2 && SVC_FLAG_L7LOADBALANCER */
	};
	/* For the service frontend, count denotes number of service backend
	 * slots (otherwise zero).
	 */
	__u16 count;
	__u16 rev_nat_index;	/* Reverse NAT ID in lb4_reverse_nat */
	__u8 flags;
	__u8 flags2;
	__u8  pad[2];
};

#endif /* __TARGET_ARCH_x86 */