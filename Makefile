# sudo apt-get install rpm
# gem install fpm
# go get github.com/mitchellh/gox

GOPATH?=$(HOME)/go

README.md: *.go
	[ -e $(GOPATH)/bin/godocdown ] || env -C / go install github.com/robertkrimen/godocdown/godocdown@latest
	$(GOPATH)/bin/godocdown >README.md~
	mv README.md~ README.md

pkgname:=$(shell basename `realpath .`)
gitdesc:=$(subst -, ,$(shell git describe --tags HEAD))
version:=$(subst v,,$(shell git tag --points-at $(word 1,$(gitdesc)) --sort=-v:refname | head -n1))$(if $(word 2,$(gitdesc)),.$(word 2,$(gitdesc)),)
iteration:=$(word 3,$(gitdesc))
iteration:=$(if $(iteration),$(iteration),1)
.PHONY: dist
dist:
	mkdir -p dist
	go generate
	go test
	go test -race
	go install
	-rm -f "dist/$(pkgname)"*
	if which upx; then upx $(GOPATH)/bin/$(pkgname); fi
	set -e; cd dist; for type in deb rpm tar; do TAR_OPTIONS="--owner=0 --group=0" fpm -t $$type -s dir -n $(pkgname) -v $(version) --iteration $(iteration) --prefix /usr/bin -C $(GOPATH)/bin $(pkgname); done
	mv dist/$(pkgname).tar dist/$(pkgname)-$(version)-$(iteration).tar
	bzip2 -v -f dist/*.tar
	gox -osarch=linux/arm -output="dist/$(pkgname)-linux-arm"
	set -e; cd dist; TAR_OPTIONS="--owner=0 --group=0" fpm -t deb --architecture armhf -s dir -n $(pkgname) -v $(version) --iteration $(iteration) --prefix /usr/bin $(pkgname)-linux-arm=$(pkgname)
	ls -l dist
