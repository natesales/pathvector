all: prepare bird gortr

clean:
	rm -rf build
	rm -rf *.rpm

prepare:
	mkdir -p build

bird:
	if [ ! -d "build/bird/" ] ; then git clone https://gitlab.nic.cz/labs/bird -b v2.0.8 build/bird/ ; fi
	cd build/bird && autoreconf
	cd build/bird && ./configure
	cd build/bird && sed -i 's/^LDFLAGS=.*/& -static/' Makefile
	cd build/bird && make
	nfpm package --packager rpm --target . --config bird-nfpm.yml

gortr:
	if [ ! -d "build/gortr/" ] ; then git clone https://github.com/cloudflare/gortr build/gortr/ ; fi
	cd build/gortr && CGO_ENABLED=0 go build cmd/gortr/gortr.go
	nfpm package --packager rpm --target . --config gortr-nfpm.yml
