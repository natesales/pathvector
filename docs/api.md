# API

Pathvector exposes an API for control and monitoring of the service. It doesn't authenticate requests, so you should protect the API endpoint behind isolated network or reverse proxy if you want it to listen on something other than loopback (the default).

## Routes
### `/show`

returns the output of `birdc show protocols`

CLI: `pathvector exec show`

### `/reload`

reloads the current configuration. If the ASN parameter is set to zero, all networks will be updated, otherwise only networks with provided ASN will be updated.

CLI: `pathvector exec reload [-a ASN]`

