package bird

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBirdConn(t *testing.T) {
	unixSocket := "test.sock"

	// Delete socket
	t.Log("Removing existing socket")
	_ = os.Remove(unixSocket)

	go func() {
		time.Sleep(time.Millisecond * 10) // Wait for the server to start
		resp, _, err := RunCommand("bird command test\n", unixSocket)
		assert.Nil(t, err)

		// Print bird output as multiple lines
		for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
			t.Logf("BIRD response (multiline): %s", line)
		}
	}()

	t.Log("Starting fake BIRD socket server")
	l, err := net.Listen("unix", unixSocket)
	assert.Nil(t, err)

	defer l.Close()
	t.Logf("Accepting connection on %s", unixSocket)
	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("0001 Fake BIRD response 1\n"))
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, err := conn.Read(buf[:])
	assert.Nil(t, err)
	assert.Equal(t, "bird command test\n", string(buf[:n]))

	_, err = conn.Write([]byte("0001 Fake BIRD response 2\n"))
	assert.Nil(t, err)
}

func TestBirdProtocolParseOne(t *testing.T) {
	p, err := ParseProtocol(`
static4    Static     master4    up     2023-03-15 19:18:50
  Channel ipv4
	State:          UP
	Table:          master4
	Preference:     200
	Input filter:   ACCEPT
	Output filter:  REJECT
	Routes:         3 imported, 2 exported, 1 preferred
	Route change stats:     received   rejected   filtered    ignored   accepted
	  Import updates:              1          0          0          0          1
	  Import withdraws:            0          0        ---          0          0
	  Export updates:              0          0          0        ---          0
	  Export withdraws:            0        ---        ---        ---          0
`)
	assert.Nil(t, err)
	assert.Equal(t, "static4", p.Name)
	assert.Equal(t, "Static", p.Proto)
	assert.Equal(t, "master4", p.Table)
	assert.Equal(t, "up", p.State)
	assert.Equal(t, "2023-03-15 19:18:50", p.Since)
	assert.Equal(t, "", p.Info)

	assert.Equal(t, 3, p.Routes.Imported)
	assert.Equal(t, 2, p.Routes.Exported)
	assert.Equal(t, 1, p.Routes.Preferred)

	p, err = ParseProtocol(`EXAMPLE_AS65522_v6 BGP        ---        up     2023-03-26 03:53:56  Established   
  BGP state:          Established
    Neighbor address: 2001:db8::1
    Neighbor AS:      65522
    Local AS:         65511
    Neighbor ID:      192.168.1.2
    Local capabilities
      Multiprotocol
        AF announced: ipv6
      Route refresh
      Graceful restart
      4-octet AS numbers
      Enhanced refresh
      Long-lived graceful restart
    Neighbor capabilities
      Multiprotocol
        AF announced: ipv6
      Route refresh
      Graceful restart
      4-octet AS numbers
      Enhanced refresh
      Long-lived graceful restart
    Session:          external AS4
    Source address:   2001:db8::1
    Hold timer:       212.093/240
    Keepalive timer:  36.625/80
  Channel ipv6
    State:          UP
    Table:          master6
    Preference:     100
    Input filter:   (unnamed)
    Output filter:  (unnamed)
    Import limit:   300000
      Action:       disable
    Routes:         176493 imported, 0 filtered, 2 exported, 175609 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:       24624786          0          0    1059583   23565203
      Import withdraws:      1469476          0        ---      12109    1457367
      Export updates:       32885258   14165213   18720043        ---          2
      Export withdraws:      1469649        ---        ---        ---          0
    BGP Next hop:   2001:db8::1 fe80::face:0:1
`)
	assert.Nil(t, err)
	assert.Equal(t, "EXAMPLE_AS65522_v6", p.Name)
	assert.Equal(t, "BGP", p.Proto)
	assert.Equal(t, "---", p.Table)
	assert.Equal(t, "up", p.State)
	assert.Equal(t, "2023-03-26 03:53:56", p.Since)
	assert.Equal(t, "Established", p.Info)
	assert.Equal(t, 176493, p.Routes.Imported)
	assert.Equal(t, 0, p.Routes.Filtered)
	assert.Equal(t, 2, p.Routes.Exported)

	assert.Equal(t, "2001:db8::1", p.BGP.NeighborAddress)
	assert.Equal(t, 65522, p.BGP.NeighborAS)
	assert.Equal(t, 65511, p.BGP.LocalAS)
	assert.Equal(t, "192.168.1.2", p.BGP.NeighborID)
}

func TestBirdParseProtocols(t *testing.T) {
	protocols, err := ParseProtocols(`
BIRD 2.0.9 ready.
Name       Proto      Table      State  Since         Info
static4    Static     master4    up     2023-03-15 19:18:50
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         1 imported, 0 exported, 0 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              1          0          0          0          1
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

static6    Static     master6    up     2023-03-15 19:18:50
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         2 imported, 0 exported, 2 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              2          0          0          0          2
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

default4   Static     master4    up     2023-03-15 19:18:50
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         1 imported, 0 exported, 1 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              1          0          0          0          1
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

default6   Static     master6    up     2023-03-15 19:18:50
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         1 imported, 0 exported, 1 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              1          0          0          0          1
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

device1    Device     ---        up     2023-03-15 19:18:50

direct1    Direct     ---        up     2023-03-15 19:18:50
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     240
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         5 imported, 0 exported, 5 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              5          0          0          0          5
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     240
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         4 imported, 0 exported, 4 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              4          0          0          0          4
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

kernel1    Kernel     master4    up     2023-03-15 19:18:50
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     10
   Input filter:   ACCEPT
   Output filter:  (unnamed)
   Routes:         0 imported, 935552 exported, 0 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              0          0          0          0          0
	 Import withdraws:            0          0        ---          0          0
	 Export updates:       76035183          0         10        ---   76035173
	 Export withdraws:      5327168        ---        ---        ---    5327168

kernel2    Kernel     master6    up     2023-03-15 19:18:50
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     10
   Input filter:   ACCEPT
   Output filter:  (unnamed)
   Routes:         0 imported, 176497 exported, 0 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              0          0          0          0          0
	 Import withdraws:            0          0        ---          0          0
	 Export updates:       41143607          0          9        ---   41143598
	 Export withdraws:      1762596        ---        ---        ---    1762596

null4      Static     master4    up     2023-03-15 19:18:50
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         1 imported, 0 exported, 1 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              1          0          0          0          1
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

null6      Static     master6    up     2023-03-15 19:18:50
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     200
   Input filter:   ACCEPT
   Output filter:  REJECT
   Routes:         1 imported, 0 exported, 1 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:              1          0          0          0          1
	 Import withdraws:            0          0        ---          0          0
	 Export updates:              0          0          0        ---          0
	 Export withdraws:            0        ---        ---        ---          0

rpki1      RPKI       ---        start  2023-03-15 19:18:50  Transport-Error
 Cache server:     rpki.example.com
 Status:           Transport-Error
 Transport:        Unprotected over TCP
 Protocol version: 1
 Session ID:       ---
 Serial number:    ---
 Last update:      ---
 Refresh timer   : ---
 Retry timer     : 80.239/90
 Expire timer    : ---
 Channel roa4
   State:          DOWN
   Table:          rpki4
   Preference:     100
   Input filter:   ACCEPT
   Output filter:  REJECT
 Channel roa6
   State:          DOWN
   Table:          rpki6
   Preference:     100
   Input filter:   ACCEPT
   Output filter:  REJECT

AS112_AS112_v4 BGP        ---        start  2023-03-15 19:18:50  Connect       Socket: No route to host
 BGP state:          Connect
   Neighbor address: 192.0.2.4
   Neighbor AS:      112
   Local AS:         65530
   Last error:       Socket: No route to host
 Channel ipv4
   State:          DOWN
   Table:          master4
   Preference:     100
   Input filter:   (unnamed)
   Output filter:  (unnamed)
   Import limit:   2
	 Action:       disable

EXAMPLE_AS65522_v4 BGP        ---        up     2023-03-26 03:53:51  Established
 BGP state:          Established
   Neighbor address: 192.168.1.2
   Neighbor AS:      65522
   Local AS:         34553
   Neighbor ID:      192.168.1.2
   Local capabilities
	 Multiprotocol
	   AF announced: ipv4
	 Route refresh
	 Graceful restart
	 4-octet AS numbers
	 Enhanced refresh
	 Long-lived graceful restart
   Neighbor capabilities
	 Multiprotocol
	   AF announced: ipv4
	 Route refresh
	 Graceful restart
	 4-octet AS numbers
	 Enhanced refresh
	 Long-lived graceful restart
   Session:          external AS4
   Source address:   192.168.1.87
   Hold timer:       231.382/240
   Keepalive timer:  32.661/80
 Channel ipv4
   State:          UP
   Table:          master4
   Preference:     100
   Input filter:   (unnamed)
   Output filter:  (unnamed)
   Import limit:   1000000
	 Action:       disable
   Routes:         935534 imported, 0 filtered, 1 exported, 718626 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:       46770469          0          0    4301505   42468964
	 Import withdraws:      3974412          0        ---      37444    3936968
	 Export updates:       60645233   33113003   27532229        ---          1
	 Export withdraws:      3889793        ---        ---        ---          0
   BGP Next hop:   192.168.1.87

EXAMPLE_AS65522_v6 BGP        ---        up     2023-03-26 03:53:56  Established
 BGP state:          Established
   Neighbor address: 2001:db8::1
   Neighbor AS:      65522
   Local AS:         65511
   Neighbor ID:      192.168.1.2
   Local capabilities
	 Multiprotocol
	   AF announced: ipv6
	 Route refresh
	 Graceful restart
	 4-octet AS numbers
	 Enhanced refresh
	 Long-lived graceful restart
   Neighbor capabilities
	 Multiprotocol
	   AF announced: ipv6
	 Route refresh
	 Graceful restart
	 4-octet AS numbers
	 Enhanced refresh
	 Long-lived graceful restart
   Session:          external AS4
   Source address:   2001:db8::1
   Hold timer:       212.093/240
   Keepalive timer:  36.625/80
 Channel ipv6
   State:          UP
   Table:          master6
   Preference:     100
   Input filter:   (unnamed)
   Output filter:  (unnamed)
   Import limit:   300000
	 Action:       disable
   Routes:         176493 imported, 0 filtered, 2 exported, 175609 preferred
   Route change stats:     received   rejected   filtered    ignored   accepted
	 Import updates:       24624786          0          0    1059583   23565203
	 Import withdraws:      1469476          0        ---      12109    1457367
	 Export updates:       32885258   14165213   18720043        ---          2
	 Export withdraws:      1469649        ---        ---        ---          0
   BGP Next hop:   2001:db8::1 fe80::face:0:1
`)
	assert.Nil(t, err)
	assert.Len(t, protocols, 14)

	protocols, err = ParseProtocols(`
BIRD 2.13 ready.
Name       Proto      Table      State  Since         Info
device1    Device     ---        up     21:26:25.230  

direct1    Direct     ---        down   21:26:25.230  
  Channel ipv4
    State:          DOWN
    Table:          master4
    Preference:     240
    Input filter:   ACCEPT
    Output filter:  REJECT
  Channel ipv6
    State:          DOWN
    Table:          master6
    Preference:     240
    Input filter:   ACCEPT
    Output filter:  REJECT

kernel1    Kernel     master4    up     21:26:25.230  
  Channel ipv4
    State:          UP
    Table:          master4
    Preference:     10
    Input filter:   ACCEPT
    Output filter:  ACCEPT
    Routes:         0 imported, 0 exported, 0 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              0          0          0          0          0
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0

kernel2    Kernel     master6    up     21:26:25.230  
  Channel ipv6
    State:          UP
    Table:          master6
    Preference:     10
    Input filter:   ACCEPT
    Output filter:  ACCEPT
    Routes:         0 imported, 0 exported, 0 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              0          0          0          0          0
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0

static1    Static     master4    up     21:26:25.230  
  Channel ipv4
    State:          UP
    Table:          master4
    Preference:     200
    Input filter:   ACCEPT
    Output filter:  REJECT
    Routes:         0 imported, 0 exported, 0 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              0          0          0          0          0
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0

`)
	assert.Nil(t, err)
	assert.Len(t, protocols, 5)
}

func TestBirdProtocolParseRoutes(t *testing.T) {
	for _, tc := range []struct {
		In     string
		Routes *Routes
	}{
		{"Routes:         176493 imported, 0 filtered, 2 exported, 175609 preferred", &Routes{Imported: 176493, Filtered: 0, Exported: 2, Preferred: 175609}},
		{"Routes:         1 imported, 0 exported, 0 preferred", &Routes{Imported: 1, Filtered: -1, Exported: 0, Preferred: 0}},
	} {
		routes, err := parseRoutes(tc.In)
		assert.Nil(t, err)
		assert.Equal(t, tc.Routes, routes)
	}
}
