# API

Pathvector exposes an API for control and monitoring of the service. It doesn't authenticate requests, so you should protect the API endpoint behind an isolated network or reverse proxy if you want it to listen on something other than loopback (the default).

## Routes
### `/show`
Returns the output of `birdc show protocols`, CLI equivalent `pathvector exec show`
### `/reload`
Reloads the current configuration. If the ASN parameter is set to zero, all networks will be updated, otherwise only networks with provided ASN will be updated. CLI equivalent `pathvector exec reload [-a ASN]`
