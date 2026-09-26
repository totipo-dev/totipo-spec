package graph

// Publication is a single-vault, single-local-key workflow model. The booleans
// are results of assertion/provenance validation and durable publication, not
// new wire fields. Production storage and crash recovery are outside this model.
type Publication struct {
	DeviceID, PublicKey      string
	advertised, tokenDurable bool
}

func (p *Publication) Advertise(deviceID, publicKey string, valid, verified, durable bool) {
	if deviceID == p.DeviceID && publicKey == p.PublicKey && valid && verified && durable {
		p.advertised = true
	}
}
func (p *Publication) PublishToken(durable bool) { p.tokenDurable = durable }
func (p *Publication) ReportSuccess() bool       { return p.advertised && p.tokenDurable }
