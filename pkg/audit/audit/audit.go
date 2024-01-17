package audit

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/audit/policy"
	"github.com/natesales/pathvector/pkg/config"
)

type Policy struct {
	ID        int
	ShortName string
	Message   string
	AppliesTo func(*config.Config, *config.Peer) bool
}

type PolicyGroup struct {
	ID        int
	ShortName string
	Message   string
	AppliesTo func(*config.Config, *config.Peer) bool
	Policies  []*Policy
	Children  []*PolicyGroup
}

type Alert struct {
	Policy *Policy
	Peer   *config.Peer
}

// Run checks if a policy applies to a peer and if so, records it
func (p *Policy) Run(c *config.Config, peer *config.Peer) {
	if p.AppliesTo(c, peer) {
		log.Warnf("AS%d: %s", *peer.ASN, message)
	}
}

// Validate traverses the policy tree to validates all policies
func (g *PolicyGroup) Validate() error {
	ids := make(map[int]bool)
	return g.validate(ids)
}

func (g *PolicyGroup) validate(ids map[int]bool) error {
	for _, policy := range g.Policies {
		if ids[policy.ID] {
			return fmt.Errorf("policy ID %d is not unique", policy.ID)
		}
		ids[policy.ID] = true
	}
	for _, childGroup := range g.Children {
		if err := childGroup.validate(ids); err != nil {
			return err
		}
	}
	return nil
}

// Run checks if a policy group applies to a peer and if so, runs all policies and child groups
func (g *PolicyGroup) Run(c *config.Config, peer *config.Peer) {
	if g.AppliesTo(c, peer) {
		for _, childPolicy := range g.Policies {
			childPolicy.Run(c, peer)
		}

		for _, childGroup := range g.Children {
			childGroup.Run(c, peer)
		}
	}
}

// Check runs all audit checks
func Check(c *config.Config) []Alert {
	var alerts []Alert

	for _, peer := range c.Peers {
		// TODO: Use alerts
		policy.RootPolicyGroup.Run(c, peer)
	}

	fmt.Printf("%+v\n", alerts)

	return alerts
}
