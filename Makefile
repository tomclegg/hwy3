# sudo apt-get install rpm
# sudo gem install fpm

README.md: *.go
	go get github.com/robertkrimen/godocdown/godocdown
	$(GOPATH)/bin/godocdown >README.md~
	mv README.md~ README.md

pkgname:=$(shell basename `realpath .`)
minor:=0.1
commitdate:=$(shell git log --first-parent --max-count=1 --format=format:%ci | tr -d - | head -c8)
commitabbrev:=$(shell git log --first-parent --max-count=1 --format=format:%h)
.PHONY: dist
dist:
	go get github.com/alecthomas/gometalinter
	mkdir -p dist
	go test
	go test -race
	gometalinter --deadline=10s
	go install
	rm "dist/$(pkgname)"*
	set -e; cd dist; for type in deb rpm tar; do TAR_OPTIONS="--owner=0 --group=0" fpm -t $$type -s dir -n $(pkgname) -v $(minor).$(commitdate) --iteration $(commitabbrev) --prefix /usr/bin -C $(GOPATH)/bin $(pkgname); done
	mv dist/$(pkgname).tar dist/$(pkgname)-$(minor).$(commitdate)-$(commitabbrev).tar
	bzip2 -v -f dist/*.tar
	ls -l dist
