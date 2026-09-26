package main

import "totipo/conformance/internal/vectors"

// Known answers transcribed from RFC 6238 Appendix B, with Appendix A's
// algorithm-specific key lengths. Never derive expectations using our consumer.
func totpCases() {
	times := []uint64{59, 1111111109, 1111111111, 1234567890, 2000000000, 20000000000}
	counters := []uint64{1, 37037036, 37037037, 41152263, 66666666, 666666666}
	counterHex := []string{"0000000000000001", "00000000023523ec", "00000000023523ed", "000000000273ef07", "0000000003f940aa", "0000000027bc86aa"}
	for _, v := range []struct {
		name      string
		algorithm byte
		secret    string
		codes     []string
	}{
		{"sha1", 1, "12345678901234567890", []string{"94287082", "07081804", "14050471", "89005924", "69279037", "65353130"}},
		{"sha256", 2, "12345678901234567890123456789012", []string{"46119246", "68084774", "67062674", "91819424", "90698825", "77737706"}},
		{"sha512", 3, "1234567890123456789012345678901234567890123456789012345678901234", []string{"90693936", "25091201", "99943326", "93441116", "38618901", "47863826"}},
	} {
		x := &vectors.TOTP{Source: "https://www.rfc-editor.org/rfc/rfc6238.html#appendix-B", Notes: "RFC 6238 Appendix B known answers; algorithm-specific 20/32/64-byte ASCII secrets from Appendix A. Counter = floor((unix_time_seconds - t0) / period), encoded as u64be. Public test material.", Algorithm: v.algorithm, Digits: 8, Period: 30, T0: 0, Secret: hx([]byte(v.secret))}
		for i, tm := range times {
			x.Rows = append(x.Rows, vectors.TOTPRow{UnixSeconds: tm, Counter: counters[i], CounterHex: counterHex[i], Code: v.codes[i]})
		}
		add(vectors.Case{ID: "v1.totp.rfc6238-" + v.name + ".001", Operation: "totp", Expected: "PASS", TOTP: x}, "bytes", "40", "56")
	}
}
