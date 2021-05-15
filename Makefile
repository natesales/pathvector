all: prepare bird gortr bgpq4 wireframe manifest swix

clean:
	rm -rf build
	rm -rf *.swix

prepare:
	mkdir -p build

bird:
	if [ ! -d "build/bird/" ] ; then git clone https://gitlab.nic.cz/labs/bird -b v2.0.8 build/bird/ ; fi
	cd build/bird && autoreconf
	cd build/bird && ./configure
	cd build/bird && sed -i 's/^LDFLAGS=.*/& -static/' Makefile
	cd build/bird && make
	nfpm package --packager rpm --target build/ --config bird-nfpm.yml

gortr:
	if [ ! -d "build/gortr/" ] ; then git clone https://github.com/cloudflare/gortr build/gortr/ ; fi
	cd build/gortr && CGO_ENABLED=0 go build cmd/gortr/gortr.go
	nfpm package --packager rpm --target build/ --config gortr-nfpm.yml

bgpq4:
	if [ ! -d "build/bgpq4/" ] ; then git clone https://github.com/bgp/bgpq4 build/bgpq4/ ; fi
	cd build/bgpq4 && ./bootstrap
	cd build/bgpq4 && LDFLAGS=-static ./configure
	cd build/bgpq4 && make
	nfpm package --packager rpm --target build/ --config bgpq4-nfpm.yml

wireframe:
	if [ ! -d "build/wireframe/" ] ; then git clone https://github.com/natesales/wireframe build/wireframe/ ; fi
	cd build/wireframe && CGO_ENABLED=0 go build
	nfpm package --packager rpm --target build/ --config wireframe-nfpm.yml

manifest:
	echo "format: 1" > build/manifest.txt
	echo "primaryRpm: $(shell ls wireframe*.rpm)" >> build/manifest.txt
	for f in $$(ls *.rpm); do echo "$$f-sha1: $$(sha1sum $$f | cut -d " " -f 1)"; done >> build/manifest.txt

swix:
	zip wireframe-bundle-$$(date +%m-%d-%Y).swix build/manifest.txt build/*.rpm
