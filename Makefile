NAME        := denote

prefix      ?= /usr/local
exec_prefix ?= $(prefix)
bindir      ?= $(exec_prefix)/bin
srcdir      ?= ./app/src

targetdir   := ./target
target      := $(targetdir)/$(NAME)
bindestdir  := $(DESTDIR)$(bindir)

all: build

build:
	# ugly fix :(
	(cd $(srcdir); go build -o ../../$(target) ".")

installdirs:
	install -d $(bindestdir)/

install: installdirs
	install $(target) $(bindestdir)/

uninstall:
	rm -f $(bindestdir)/$(NAME)

clean:
	rm -rf $(targetdir)/

docker:
	docker build -t $(NAME) -f ./app/Dockerfile "."

clean-docker:
	docker rmi $(NAME)
