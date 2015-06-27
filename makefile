export CC=gcc
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
export INC=-I $(ROOT_DIR)/lib

all: client server

client:
	$(MAKE) -C client

server:
	$(MAKE) -C server

.PHONY: all clean client server

clean:
	$(MAKE) -C client clean
	$(MAKE) -C server clean
