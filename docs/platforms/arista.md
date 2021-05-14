# Arista

In theory, it's possible to install RPM packages on EOS, but due to version differences and very old Fedora/CentOS releases, it's a lot easier to just build a statically linked binary for the switch and copy it over.

## Preparing the directory structure

The easiest way to get started with Wireframe on Arista EOS is to create a simple directory structure and rc script on the mounted flash directory. It's also possible to create a ProcMgr service in `/etc/ProcMgr.d/` instead.  

```bash
mkdir /mnt/flash/{bin,etc}
cat <<EOF > /mnt/flash/rc.eos
#!/bin/bash

touch /run/bird.ctl
cp /mnt/drive/bin/* /bin/
cp /mnt/drive/etc/* /etc/
bird # BIRD will fork itself into the background by default
EOF
```

## Installing BIRD

To compile statically linked BIRD binaries, first clone the repo from `https://gitlab.nic.cz/labs/bird` and follow their build instructions with one notable exception: before running `make`, add the `-static` flag to `LDFLAGS` in the `Makefile` (`sed -i '/^LDFLAGS=.*/a LDFLAGS := -static' Makefile`).

## Installing GoRTR

GoRTR releases are currently dynamically linked, so we need to compile them with CGO disabled after cloning the [repo](https://github.com/cloudflare/gortr): `CGO_ENABLED=0 go build cmd/gortr/gortr.go`

## Installing Wireframe

[Wireframe releases](https://github.com/natesales/wireframe/releases/) are already statically linked binaries, so it's as easy as downloading the latest binary release, extracting it, and copying the resulting `wireframe` binary to `/mnt/flash/bin/`. Make sure to check your switch architecture to download the correct binary (`bash uname -a` from the EOS CLI).

## Copy the binaries

Using `scp` or a USB drive, copy each required binary to `/mnt/flash/bin/` on the Arista.
