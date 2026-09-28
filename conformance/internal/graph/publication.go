package graph

// Publication is a single-vault, single-local-key workflow model. The booleans
// represent correct construction/signing and local publication acknowledgement, not
// new wire fields. Production storage and crash recovery are outside this model.
type Publication struct {
	DeviceID, PublicKey           string
	advertised, tokenAcknowledged bool
}

func (p *Publication) Advertise(deviceID, publicKey string, valid, verified, acknowledged bool) {
	if deviceID == p.DeviceID && publicKey == p.PublicKey && valid && verified && acknowledged {
		p.advertised = true
	}
}
func (p *Publication) PublishToken(acknowledged bool) { p.tokenAcknowledged = acknowledged }
func (p *Publication) ReportSuccess() bool            { return p.advertised && p.tokenAcknowledged }
