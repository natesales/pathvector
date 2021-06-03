package main

import (
	"os/exec"

	"github.com/joomcode/errorx"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// configureInterfaces applies interface configs
func configureInterfaces(config *config) {
	for ifaceName, ifaceOpts := range config.Interfaces {
		if ifaceOpts.Dummy {
			log.Infof("Creating new dummy interface: %s", ifaceName)
			linkAttrs := netlink.NewLinkAttrs()
			linkAttrs.Name = ifaceName
			newIface := &netlink.Dummy{LinkAttrs: linkAttrs}
			if err := netlink.LinkAdd(newIface); err != nil {
				log.Warn(errorx.Decorate(err, "dummy interface create"))
			}
		}

		// Get link by name
		link, err := netlink.LinkByName(ifaceName)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("found interface %s index %d", ifaceName, link.Attrs().Index)

		// Set MTU
		if ifaceOpts.MTU != 0 {
			if err := netlink.LinkSetMTU(link, int(ifaceOpts.MTU)); err != nil {
				log.Warn(errorx.Decorate(err, "set MTU on "+ifaceName))
			}
		}

		// Add addresses
		for _, addr := range ifaceOpts.Addresses {
			nlAddr, err := netlink.ParseAddr(addr)
			if err != nil {
				log.Fatal(err) // This should never happen
			}
			if err := netlink.AddrAdd(link, nlAddr); err != nil {
				log.Warn(errorx.Decorate(err, "add address to "+ifaceName))
			}
		}

		// Add interfaces to xdprtr dataplane
		if ifaceOpts.XDPRTR {
			out, err := exec.Command("xdprtrload", ifaceName).Output()
			if err != nil {
				log.Fatalf("xdprtrload: %v", err)
			}
			log.Infof("xdprtrload: " + string(out))
		}

		// Set interface status
		if ifaceOpts.Down {
			if err := netlink.LinkSetDown(link); err != nil {
				log.Fatal(errorx.Decorate(err, "set link down"))
			}
		} else {
			if err := netlink.LinkSetUp(link); err != nil {
				log.Fatal(errorx.Decorate(err, "set link down"))
			}
		}
	}
}
