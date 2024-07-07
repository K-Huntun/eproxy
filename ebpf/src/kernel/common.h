// Supplement vmlinux
#ifndef _LINUX_COMMON_H
#define _LINUX_COMMON_H

// 保证4的倍数对其
struct lb4_key {
	__be32 address;		/* Service virtual IPv4 address */
	__be16 dport;		/* L4 port filter, if unset, all ports apply */
	__u8 proto;		/* L4 protocol, currently not used (set to 0) */
	__u8 pad;
};

struct lb4_backend {
	__be32 address;		/* Service endpoint IPv4 address */
	__be16 port;		/* L4 port filter */
	__u8 proto;		/* L4 protocol, currently not used (set to 0) */
	__u8 pad;
};

struct lb4_service {
	__u16 service_id;
	__u16 count;
};

#ifndef __section_maps
#define __section_maps			SEC("maps")
#endif

#ifndef __section_maps_btf
#define __section_maps_btf		SEC(".maps")
#endif

#endif /* __TARGET_ARCH_x86 */