package vectors

import (
	"bytes"
	"fmt"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
)

func runSignatureContext(c Case) error {
	x := c.Context
	a, e := unhex(x.A)
	if e != nil {
		return e
	}
	b, e := unhex(x.B)
	if e != nil {
		return e
	}
	if len(a) != 32 || len(b) != 32 || bytes.Equal(a, b) || x.UnderA != "VERIFIED" || x.UnderB != "REJECTED" {
		return fmt.Errorf("invalid vault contexts/expectations")
	}
	pub, e := unhex(c.PublicKey)
	if e != nil {
		return e
	}
	u, e := c.Input.Unsigned()
	if e != nil {
		return e
	}
	if e = equalHex("unsigned semantic", x.Unsigned, u); e != nil {
		return e
	}
	for _, v := range []struct {
		context []byte
		want    string
	}{{a, x.UnderA}, {b, x.UnderB}} {
		k := cryptov1.Keys{SignatureContext: v.context}
		if got := k.Provenance(*c.Input, pub); got != v.want {
			return fmt.Errorf("context provenance: got %s want %s", got, v.want)
		}
	}
	return nil
}
func runPublication(c Case) error {
	x := c.Publication
	if x.DeviceID == "" || x.PublicKey == "" {
		return fmt.Errorf("missing local identity")
	}
	p := graph.Publication{DeviceID: x.DeviceID, PublicKey: x.PublicKey}
	checks := 0
	for _, event := range x.Events {
		switch event.Action {
		case "advertise":
			p.Advertise(event.DeviceID, event.PublicKey, event.Valid, event.Verified, event.Durable)
		case "publish-token":
			p.PublishToken(event.Durable)
		case "report-success":
			if event.Success == nil || p.ReportSuccess() != *event.Success {
				return fmt.Errorf("publication success gate mismatch")
			}
			checks++
		case "restart-workflow":
			p = graph.Publication{DeviceID: x.DeviceID, PublicKey: x.PublicKey}
		default:
			return fmt.Errorf("unknown publication event")
		}
	}
	if checks == 0 {
		return fmt.Errorf("missing publication assertions")
	}
	return nil
}
