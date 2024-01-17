package policy

import (
	"github.com/natesales/pathvector/pkg/audit"
	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

func init() {
	if err := RootPolicyGroup.Validate(); err != nil {
		panic(err)
	}
}

var (
	RootPolicyGroup = &audit.PolicyGroup{
		ID:        0,
		ShortName: "ROOT",
		Message:   "Root policy group",
		AppliesTo: func(c *config.Config, peer *config.Peer) bool { return true },
		Policies:  []*audit.Policy{missingRole},
		Children:  []*audit.PolicyGroup{groupEBGP, groupTier1},
	}

	missingRole = &audit.Policy{ID: 1, ShortName: "BGP-ROLE", Message: "Missing RFC 9234 BGP role", AppliesTo: func(c *config.Config, peer *config.Peer) bool { return peer.Role == nil }}

	groupEBGP = &audit.PolicyGroup{
		ID:        1,
		ShortName: "EBGP",
		Message:   "eBGP peer missing required policies",
		AppliesTo: func(c *config.Config, peer *config.Peer) bool {
			// If localAS != remoteAS and remoteAS is not private
			return *peer.ASN != c.ASN && !util.IsPrivateASN(uint32(*peer.ASN))
		},
		Policies: []*audit.Policy{
			{
				ID: 2, ShortName: "RPKI-FILTER", Message: "Public eBGP peer lacks RPKI filtering",
				AppliesTo: func(c *config.Config, peer *config.Peer) bool {
					return !*peer.FilterRPKI
				},
			},
			{
				ID: 3, ShortName: "EBGP-PREFIX-LIMIT", Message: "Public eBGP peer lacks max prefix filtering",
				AppliesTo: func(c *config.Config, peer *config.Peer) bool {
					return !*peer.FilterMaxPrefix
				},
			},
			{
				ID: 4, ShortName: "EBGP-BOGON-FILTER", Message: "Public eBGP peer lacks bogon route filtering", AppliesTo: func(c *config.Config, peer *config.Peer) bool {
					return !*peer.FilterBogonRoutes
				},
			},
			{
				ID: 5, ShortName: "EBGP-BOGON-AS-FILTER", Message: "Public eBGP peer lacks bogon ASN filtering", AppliesTo: func(c *config.Config, peer *config.Peer) bool {
					return !*peer.FilterBogonASNs
				},
			},
		},
	}

	groupTier1 = &audit.PolicyGroup{
		ID:        2,
		ShortName: "TIER-1",
		Message:   "Tier 1 peer missing required policies",
		AppliesTo: func(c *config.Config, peer *config.Peer) bool {
			// If the peer is a tier 1 and we aren't
			return util.Contains(c.TransitASNs, uint32(*peer.ASN)) && util.Contains(c.TransitASNs, uint32(c.ASN))
		},
		Policies: []*audit.Policy{},
	}
)
