//go:build ignore
#include "vmlinux.h"
#include "socket.h"
#include "bpf_helpers.h"
#include "bpf_endian.h"
#include "common.h"

#define CILIUM_LB_REV_NAT_MAP_MAX_ENTRIES	65536
#define CILIUM_LB_SERVICE_MAP_MAX_ENTRIES	65536
#define CILIUM_LB_BACKENDS_MAP_MAX_ENTRIES	65536
#define CILIUM_LB_AFFINITY_MAP_MAX_ENTRIES	65536
#define CILIUM_LB_REV_NAT_MAP_MAX_ENTRIES	65536
#define CILIUM_LB_MAGLEV_MAP_MAX_ENTRIES	65536
#define CONDITIONAL_PREALLOC 0

char __license[] SEC("license") = "Dual MIT/GPL";

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, struct lb4_key);
	__type(value, struct lb4_service);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, CILIUM_LB_SERVICE_MAP_MAX_ENTRIES);
} eproxy_lb4_services __section_maps_btf;

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, __u32);
	__type(value, struct lb4_backend);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, CILIUM_LB_BACKENDS_MAP_MAX_ENTRIES);
} eproxy_lb4_backends __section_maps_btf;

SEC("cgroup/connect4")
int connect4(struct bpf_sock_addr *ctx) {
    int ret = 1; /* OK value */
    if (ctx->type != SOCK_STREAM && ctx->type != SOCK_DGRAM) {
        bpf_printk("unkonw socket type");
        return ret;
    }
    struct lb4_key key = {};
    key.address = ctx->user_ip4;
    key.dport = ctx->user_port;
    key.backend_slot = 0;
    key.proto = 0;
    bpf_printk("before user ip %lu, port:%lu\n",ctx->user_ip4, ctx->user_port);
    struct lb4_service* value= bpf_map_lookup_elem(&eproxy_lb4_services ,&key);
    if (value == NULL){
        return 1;
    }

    __u16 count = value->count;
    if (count == 0){
        return 1;
    }
    __u16 index = (bpf_get_prandom_u32() % count)+1;
    __u32 blackend_id = value->service_id << 16 | index;
    struct lb4_backend* end_value = bpf_map_lookup_elem(&eproxy_lb4_backends ,&blackend_id);
    if (end_value == NULL){
        return 1;
    }
    ctx->user_ip4 = end_value->address;
    ctx->user_port = end_value->port;
    bpf_printk("after user ip %lu, port:%lu\n",ctx->user_ip4, ctx->user_port);
    return 1;
}
