CLANG   ?= clang
LLC     ?= llc
OPT     ?= opt
DIS     ?= llvm-dis
RM      ?= rm
MV      ?= mv
CC      ?= gcc
MAKE    ?= make
OBJDUMP ?= llvm-objdump

HEADER := -I./headers         \
         -I./kernel

CFLAGS_KV :=					\
	-g -O2 -Wall                            \
	-emit-llvm -target bpf                  \
	-D__KERNEL__                            \
	-D__TARGET_ARCH_x86_64                  \
	-D__NATIVE_EBPF__						\
	-D__x86_64__

all: service.o

service.o: service.c
	$(CLANG) $(CFLAGS_KV) $(HEADER) -O2 -c $< -o -  | \
	$(LLC) -march=bpf -filetype=obj -o service.o
	$(MV) service.o ../output/

clean:
	$(RM) -rf ../output/service.o