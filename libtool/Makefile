.PHONY: install uninstall 

LIBPATH = $(shell pwd)
LIBTOOLS = m4-1.4.18 bison-3.0.5 flex-2.6.4

all: install

install:
	@for dir in $(LIBTOOLS); do \
        if [ -f "$(LIBPATH)/src/$$dir/configure" ]; then \
            echo "Configuring $$dir..."; \
            cd $(LIBPATH)/src/$$dir && ./configure --prefix=$(LIBPATH) > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
            echo "Building $$dir..."; \
            cd $(LIBPATH)/src/$$dir && $(MAKE) > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
            echo "Installing $$dir..."; \
            cd $(LIBPATH)/src/$$dir && $(MAKE) install > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
        else \
            echo "Building $$dir..."; \
            cd $(LIBPATH)/src/$$dir && $(MAKE) PREFIX=$(LIBPATH) > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
            echo "Installing $$dir..."; \
            cd $(LIBPATH)/src/$$dir && $(MAKE) install > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
        fi \
    done

uninstall:
	@for dir in $(LIBTOOLS); do \
        echo "Cleaning $$dir... "; \
        if [ -f "$(LIBPATH)/src/$$dir/Makefile" ]; then \
            cd $(LIBPATH)/src/$$dir && $(MAKE) clean > /dev/null; \
            if [ $$? != 0 ]; then exit 1; fi; \
        fi; \
    done
