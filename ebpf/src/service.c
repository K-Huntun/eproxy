//go:build ignore
#include "vmlinux.h"
#include "socket.h"
#include "bpf_helpers.h"
#include "bpf_endian.h"
#include "command.h"

char __license[] SEC("license") = "Dual MIT/GPL";

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, struct lb4_key);
	__type(value, struct lb4_service);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, CILIUM_LB_SERVICE_MAP_MAX_ENTRIES);
	__uint(map_flags, CONDITIONAL_PREALLOC);
} LB4_SERVICES_MAP_V2 __section_maps_btf;

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, __u32);
	__type(value, struct lb4_backend);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, CILIUM_LB_BACKENDS_MAP_MAX_ENTRIES);
	__uint(map_flags, CONDITIONAL_PREALLOC);
} LB4_BACKEND_MAP __section_maps_btf;

SEC("cgroup/connect4")
int connect4(struct bpf_sock_addr *ctx) {
    int ret = 1; /* OK value */
    if (ctx->type != SOCK_STREAM && ctx->type != SOCK_DGRAM) {
        bpf_printk("unkonw socket type");
        return ret;
    }

    __u8 ip_proto;
    switch (ctx->type) {
    case SOCK_STREAM:
        ip_proto = IPPROTO_TCP;
        break;
    case SOCK_DGRAM:
        ip_proto = IPPROTO_UDP;
        break;
    default:
        return ret;
    }
    struct lb4_key key = {};
    key.address = ctx->user_ip4;
    key.dport = ctx->user_port;
    key.backend_slot = 0;
    key.proto = 0;
    key.scope = 0;

    struct *lb4_service value= bpf_map_lookup_elem(&LB4_SERVICES_MAP_V2 ,&key);
    if (value == NULL){
        return 1;
    }

    __u16 count = value->count;
    if (count == 0){
        return 1;
    }
    __u16 index = (bpf_get_prandom_u32() % count)+1;
    __32 backend_slot = key.backend_slot;
    struct lb4_backend* value = bpf_map_lookup_elem(&LB4_BACKEND_MAP ,&backend_slot);
    if (value == NULL){
        return 1;
    }
    ctx->user_ip4 = value->address;
    ctx->user_port = value->port;
    return 1;
}
