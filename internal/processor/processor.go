package processor

import (
	"bytes"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/irr"
	"github.com/natesales/pathvector/internal/peeringdb"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
)

func Run(global *config.Global, peerName string, peerData *config.Peer) error {
	log.Printf("Processing AS%d %s", *peerData.ASN, peerName)

	// Set sanitized peer name
	peerData.ProtocolName = util.Sanitize(peerName)

	// If a PeeringDB query is required
	if *peerData.AutoImportLimits || *peerData.AutoASSet {
		log.Debugf("[%s] has auto-import-limits or auto-as-set, querying PeeringDB", peerName)

		if err := peeringdb.Run(peerData, global.PeeringDbQueryTimeout); err != nil {
			log.Debugf("[%s] %v", peerName, err)
		}
	} // end peeringdb query enabled

	// Build IRR prefix sets
	if *peerData.FilterIRR {
		if err := irr.BuildPrefixSet(peerData, global.IRRServer, global.IRRQueryTimeout); err != nil {
			return err
		}
	}

	util.PrintStructInfo(peerName, peerData)

	// Create peer file
	peerFileName := path.Join(global.CacheDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *util.Sanitize(peerName)))
	peerSpecificFile, err := os.Create(peerFileName)
	if err != nil {
		return fmt.Errorf("Create peer specific output file: %v", err)
	}

	// Render the template and write to buffer
	var b bytes.Buffer
	log.Debugf("[%s] Writing config", peerName)
	err = templating.PeerTemplate.ExecuteTemplate(&b, "peer.tmpl", &templating.ConfigWrapper{Name: peerName, Peer: *peerData, Global: *global})
	if err != nil {
		return fmt.Errorf("Execute template: %v", err)
	}

	// Reformat config and write template to file
	if _, err := peerSpecificFile.Write([]byte(templating.ReformatBirdConfig(b.String()))); err != nil {
		return fmt.Errorf("Write template to file: %v", err)
	}

	log.Debugf("[%s] Wrote config", peerName)

	return nil // nil error
}
