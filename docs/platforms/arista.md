# Arista

## Installing BIRD

In theory, it's possible to install BIRD through a yum repo, but with the extra complexity of EOS it's a lot easier to just build a statically linked binary for the switch and copy it over.

### Create the switch init directory structure

Normal EOS services use ProcMgr to control process lifecycle with a simple but neat heartbeat system. While we could create a ProcMgr service ourselves by adding it under `/etc/ProcMgr.d/`, BIRD doesn't support the file based heartbeat method, so it's easier to just create a rc script that EOS will run on startup. Additionally, EOS doesn't persist certain directories, so we'll have to copy over the binaries from flash on boot. 

```bash
mkdir /mnt/flash/{bin,etc}
cat <<EOF > /mnt/flash/rc.eos
#!/bin/bash

touch /run/bird.ctl
cp /mnt/drive/bin/* /bin/
cp /mnt/drive/etc/* /etc/
bird
EOF
```

### Compile statically linked BIRD binaries

To compile statically linked BIRD binaries, first clone the repo from `https://gitlab.nic.cz/labs/bird` and follow their build instructions with one notable exception: before running `make`, add the `-static` flag to `LDFLAGS` in the `Makefile` (`sed -i '/^LDFLAGS=.*/a LDFLAGS := -static' Makefile`).

### Copy the binaries to the switch

Using `scp` or a USB drive, copy the `bird` and `birdc` binaries to `/mnt/flash/bin/`
