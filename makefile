export CC=gcc

all: client server

client:
	$(MAKE) -C client

server:
	$(MAKE) -C server

.PHONY: all clean client server

clean:
	$(MAKE) -C client clean
	$(MAKE) -C server clean
