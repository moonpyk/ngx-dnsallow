package config

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	LineTypeA    = `A`
	LineTypeAAAA = `AAAA`
	VerbAllow    = "allow"
)

type Entry struct {
	Dns       string `yaml:"dns"`
	Type      string `yaml:"type"`
	Mask      int    `yaml:"mask,omitempty"`
	AllowVerb string `yaml:"allowverb,omitempty"`
}

// EnsureVerb ensures a verb is given for generate-allow configurations
// defaulting to "allow" if it was configured empty
func (e *Entry) EnsureVerb() string {
	if e.AllowVerb == "" {
		return VerbAllow
	}

	return e.AllowVerb
}

// EnsureType ensures a type is given for the generate-allow configurations
// defaulting to "A" if it was configured empty
func (e *Entry) EnsureType() string {
	if len(e.Type) == 0 {
		return LineTypeA
	}

	return e.Type
}

// EnsureValidMask ensure a valid mask is given for configurations
// To avoid nginx configuration warnings, special combinations are handled :
// for A (IPv4) entries and masks length of 32 a 0 mask is returned
// for AAAA (IPv6) entries and masks length of 128 a 0 mask is returned
func (e *Entry) EnsureValidMask() int {
	lineType := e.EnsureType()

	if lineType == LineTypeA && e.Mask == 32 {
		return 0
	}

	if lineType == LineTypeAAAA && e.Mask == 128 {
		return 0
	}

	return e.Mask
}

// LookupIP looks up the host using the local resolver
func (e *Entry) LookupIP() ([]net.IP, error) {
	return net.LookupIP(e.Dns)
}

// RenderAllowLine Renders a line for the host for a generate-allow generation using [resolved] addresses
func (e *Entry) RenderAllowLine(resolved []net.IP) (string, error) {
	if len(resolved) == 0 {
		return "", errors.New("no addresses resolved")
	}

	lineType := e.EnsureType()

	switch lineType {
	case LineTypeA, LineTypeAAAA:
	default:
		return "", errors.New("invalid dns type \"" + lineType + "\"")
	}

	maskLen := e.EnsureValidMask()

	var sb strings.Builder

	for _, addr := range resolved {
		if maskLen > 0 {
			addr = addr.Mask(net.CIDRMask(maskLen, 8*len(addr)))
		}

		if addr == nil {
			continue
		}

		switch {
		case len(addr) == net.IPv4len && lineType == LineTypeA,
			len(addr) == net.IPv6len && lineType == LineTypeAAAA:
			sb.WriteString(e.EnsureVerb() + " " + addr.String())

			if maskLen > 0 {
				sb.WriteString("/" + strconv.Itoa(maskLen))
			}

			sb.WriteString("; # " + e.Dns + " [" + e.Type + "]\n")
		default:
			continue
		}
	}

	if sb.Len() == 0 {
		return "", errors.New("yielded no result")
	}

	return sb.String(), nil
}
