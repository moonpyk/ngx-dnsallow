package config

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

type Entry struct {
	Dns       string `yaml:"dns"`
	Type      string `yaml:"type"`
	Mask      int    `yaml:"mask,omitempty"`
	AllowVerb string `yaml:"allowverb,omitempty"`
}

// EnsureVerb Ensures a valid verb is given for generate-allow configurations
// defaulting to "allow" it was configured empty
func (e *Entry) EnsureVerb() string {
	if e.AllowVerb == "" {
		return "allow"
	}

	return e.AllowVerb
}

func (e *Entry) EnsureType() string {
	if len(e.Type) == 0 {
		return "A"
	}

	return e.Type
}

func (e *Entry) RenderAllowLine(addrs []net.IP) (error, string) {
	if len(addrs) == 0 {
		return errors.New("no addresses resolved"), ""
	}

	lineType := e.EnsureType()

	switch lineType {
	case "A":
	case "AAAA":
		break
	default:
		return errors.New("invalid dns type \"" + lineType + "\""), ""
	}

	maskLen := e.Mask

	if lineType == "A" && maskLen == 32 {
		maskLen = 0
	}

	if lineType == "AAAA" && maskLen == 128 {
		maskLen = 0
	}

	var sb strings.Builder

	for _, addr := range addrs {
		if maskLen > 0 {
			addr = addr.Mask(net.CIDRMask(maskLen, 8*len(addr)))
		}

		if addr == nil {
			continue
		}

		if len(addr) == net.IPv4len && lineType == "A" ||
			len(addr) == net.IPv6len && lineType == "AAAA" {
			sb.WriteString(e.EnsureVerb() + " " + addr.String())
			if maskLen > 0 {
				sb.WriteString("/" + strconv.Itoa(maskLen))
			}
			sb.WriteString("; # " + e.Dns + "[" + e.Type + "]\n")
		} else {
			continue
		}
	}

	if sb.Len() == 0 {
		return errors.New("yielded no result"), ""
	}

	return nil, sb.String()
}
