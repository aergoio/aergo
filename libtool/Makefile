.PHONY: install uninstall 

LIBPATH = $(shell pwd)
LIBTOOLS = # m4-1.4.18 bison-3.0.5 flex-2.6.4

all: install

install:
	@echo -e "\033[32m[INFO]  \033[0mInstalling dependent tools"
	@for dir in $(LIBTOOLS); do \
		if [ ! -f "$(LIBPATH)/src/$$dir/Makefile" ]; then \
			echo -e "\033[32m[INFO]  \033[0m--> Configuring $$dir"; \
			cd $(LIBPATH)/src/$$dir && ./configure --prefix=$(LIBPATH) --silent; \
			if [ $$? != 0 ]; then exit 1; fi; \
		fi; \
		echo -e "\033[32m[INFO]  \033[0m--> Installing $$dir"; \
		cd $(LIBPATH)/src/$$dir && $(MAKE) PREFIX=$(LIBPATH) install > install.log; \
		if [ $$? != 0 ]; then exit 1; fi; \
	done

uninstall:
	@echo -e "\033[32m[INFO]  \033[0mUninstalling dependent tools"
	@for dir in $(LIBTOOLS); do \
		echo -e "\033[32m[INFO]  \033[0m--> Cleaning $$dir"; \
		if [ -f "$(LIBPATH)/src/$$dir/Makefile" ]; then \
			cd $(LIBPATH)/src/$$dir && $(MAKE) clean > clean.log; \
			if [ $$? != 0 ]; then exit 1; fi; \
		fi; \
	done
