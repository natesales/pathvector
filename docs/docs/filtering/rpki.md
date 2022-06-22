# RPKI Validation

_Resource Public Key Infrastructure_ is a system to verify that an ASN is authorized to originate a route. RPKI works in
3 parts:

**Signing** a prefix involves publishing a _Route Origin Authorization (ROA)_ for each prefix that you want to
originate. ROAs contain the authorized origin ASN and a max length field to indicate how far disaggregation should be
permitted. For example, you could create a ROA for your /22 permitting announcements up to /24.

**Validation** uses the chain of trust to verify that a ROA is valid.

**Policy Enforcement** uses Validated ROA Payloads (VRPs) to influence routing decisions. These are typically
distributed from the validator(s) to routers using the *RTR (RPKI to Router)* protocol.

The global `rtr-server` option in Pathvector specifies the RTR server for the router. Enable `filter-rpki` to filter
RPKI invalid routes on a peer. The `strict-rpki` option filters prefixes that are not covered by a RPKI ROA. This is
potentially dangerous as a large portion of the Internet does not have covering ROAs.
