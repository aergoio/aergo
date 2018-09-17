# 3rd party libs

.PHONY: install uninstall libs-clean libluajit-clean

LIBPATH := $(shell pwd)

libluajit := $(LIBPATH)/lib/libluajit-5.1.a
libluajit_SRCDIR := $(LIBPATH)/src/luajit/src
libluajit_MAKEDIR := $(LIBPATH)/src/luajit

liblmdb := $(LIBPATH)/lib/liblmdb.so
liblmdb_SRCDIR := $(LIBPATH)/src/lmdb/libraries/liblmdb
liblmdb_MAKEDIR := $(liblmdb_SRCDIR)

install: $(libluajit) $(liblmdb)

$(libluajit): | $(liblua_SRCDIR)
	$(MAKE) PREFIX=$(LIBPATH) -C $(libluajit_MAKEDIR) all install

$(libluajit_SRCDIR):
	git submodule init
	git submodule update

$(liblmdb): | $(liblmdb_SRCDIR)
	$(MAKE) prefix=$(LIBPATH) -C $(liblmdb_MAKEDIR) all install

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

