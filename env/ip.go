package env

import "net/netip"

// IP parses the environment value for key as an IP address
// (v4 or v6, zones accepted), or returns fallback when unset.
func IP(key string, fallback netip.Addr) (netip.Addr, error) {
	v, ok := lookup(key)
	if !ok {
		return fallback, nil
	}
	return parseIP(key, v)
}

// parseIP is the shared parser used by IP and IPVar.
func parseIP(key, value string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Addr{}, parseError(key, "ip", value)
	}
	return addr, nil
}

// IPVar defines an IP address variable. A zero Addr default renders as
// "invalid IP"; prefer a valid default or Require().
func IPVar(key string, def netip.Addr, opts ...VarOpt) Var[netip.Addr] {
	c := applyOpts(opts)
	return Var[netip.Addr]{Key: key, Default: def, Kind: "ip", Secret: c.secret, Required: c.required, Parse: parseIP}
}
