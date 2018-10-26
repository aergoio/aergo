# 3rd party libs

.PHONY: install uninstall libs-clean libluajit-clean libluajit-build liblmdb-build

LIBPATH := $(shell pwd)

libluajit := $(LIBPATH)/lib/libluajit-5.1.a
libluajit_SRCDIR := $(LIBPATH)/src/luajit/src
libluajit_MAKEDIR := $(LIBPATH)/src/luajit

liblmdb := $(LIBPATH)/lib/liblmdb.a
liblmdb_SRCDIR := $(LIBPATH)/src/lmdb/libraries/liblmdb
liblmdb_MAKEDIR := $(liblmdb_SRCDIR)

all: libluajit-build liblmdb-build
install: $(libluajit) $(liblmdb)

libluajit-build: | $(libluajit_SRCDIR)
	$(MAKE) PREFIX=$(LIBPATH) -C $(libluajit_MAKEDIR) all

$(libluajit): $(libluajit_SRCDIR)/libluajit.a
	$(MAKE) PREFIX=$(LIBPATH) -C $(libluajit_MAKEDIR) install

$(libluajit_SRCDIR):
	git submodule init
	git submodule update

liblmdb-build: | $(liblmdb_SRCDIR)
	$(MAKE) prefix=$(LIBPATH) -C $(liblmdb_MAKEDIR) all

$(liblmdb): $(liblmdb_SRCDIR)/liblmdb.a
	$(MAKE) prefix=$(LIBPATH) -C $(liblmdb_MAKEDIR) install

$(liblmdb_SRCDIR):
	git clone https://github.com/lmdb/lmdb.git src/lmdb

uninstall: libs-clean
	rm -rf $(LIBPATH)/lib/*
	rm -rf $(LIBPATH)/include/*
	rm -rf $(LIBPATH)/bin/*
	rm -rf $(LIBPATH)/share/*

libs-clean: libluajit-clean liblmdb-clean

libluajit-clean:
	$(MAKE) PREFIX=$(LIBPATH) -C $(libluajit_MAKEDIR) clean

liblmdb-clean:
	$(MAKE) prefix=$(LIBPATH) -C $(liblmdb_MAKEDIR) clean

