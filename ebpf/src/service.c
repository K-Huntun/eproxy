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

static __always_inline __u8 get_l4_proto(struct bpf_sock_addr *ctx)
{
    if (ctx->protocol != 0) {
        return ctx->protocol;
    }

    if (ctx->type == SOCK_STREAM) {
        return 6;
    }

    if (ctx->type == SOCK_DGRAM) {
        return 17;
    }

    return 0;
}

static __always_inline int lb4_redirect(struct bpf_sock_addr *ctx)
{
    struct lb4_key key = {};
    struct lb4_service *svc;
    struct lb4_backend *backend;
    __u16 count;
    __u16 index;
    __u32 backend_id;

    if (ctx->type != SOCK_STREAM && ctx->type != SOCK_DGRAM) {
        return 1;
    }

    key.address = ctx->user_ip4;
    key.dport = ctx->user_port;
    key.proto = get_l4_proto(ctx);

    svc = bpf_map_lookup_elem(&eproxy_lb4_services, &key);
    if (svc == NULL) {
        return 1;
    }

    count = svc->count;
    if (count == 0) {
        return 1;
    }

    index = (bpf_get_prandom_u32() % count) + 1;
    backend_id = ((__u32)svc->service_id << 16) | index;
    backend = bpf_map_lookup_elem(&eproxy_lb4_backends, &backend_id);
    if (backend == NULL) {
        return 1;
    }

    ctx->user_ip4 = backend->address;
    ctx->user_port = backend->port;
    return 1;
}

SEC("cgroup/connect4")
int connect4(struct bpf_sock_addr *ctx)
{
    return lb4_redirect(ctx);
}

SEC("cgroup/sendmsg4")
int sendmsg4(struct bpf_sock_addr *ctx)
{
    return lb4_redirect(ctx);
}
