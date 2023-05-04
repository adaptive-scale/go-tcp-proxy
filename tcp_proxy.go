package proxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
)

var (
	matchid = uint64(0)
	connid  = uint64(0)
)

type TcpProxy struct {
	Logger           ColorLogger
	Listener         *net.TCPListener
	RemoteAddr       *net.TCPAddr
	LocalAddr        *net.TCPAddr
	LocalAddrString  string
	RemoteAddrString string
	Match            string
	Verbose          bool
	TLSEnabled       bool
	RootCertLocation string
	Replace          string
	UnWrapTLS        bool
	Nagle            bool
	Hex              bool
	Color            bool
}

func (t *TcpProxy) Start() error {
	matcher := t.createMatcher(t.Match)
	replacer := t.createReplacer(t.Replace)

	for {
		conn, err := t.Listener.AcceptTCP()
		if err != nil {
			t.Logger.Warn("Failed to accept connection '%s'", err)
			continue
		}
		connid++

		var p *Proxy
		if t.UnWrapTLS {
			t.Logger.Info("Unwrapping TLS")
			p = NewTLSUnwrapped(conn, t.LocalAddr, t.RemoteAddr, t.RemoteAddrString)
		} else {
			p = New(conn, t.LocalAddr, t.RemoteAddr)
		}

		if t.TLSEnabled {
			certData, err := ioutil.ReadFile(t.RootCertLocation)
			if err != nil {
				panic(fmt.Errorf("could not read certificate. err %w", err))
			}
			p.PemCert = string(certData)
		}

		p.Matcher = matcher
		p.Replacer = replacer

		p.Nagles = t.Nagle
		p.OutputHex = t.Hex
		p.Log = ColorLogger{
			Verbose:     true,
			VeryVerbose: t.Verbose,
			Prefix:      fmt.Sprintf("Connection #%03d ", connid),
			Color:       t.Color,
		}

		go p.Start()
	}
}

func (t *TcpProxy) createMatcher(match string) func([]byte) {
	if match == "" {
		return nil
	}
	re, err := regexp.Compile(match)
	if err != nil {
		t.Logger.Warn("Invalid match regex: %s", err)
		return nil
	}

	t.Logger.Info("Matching %s", re.String())
	return func(input []byte) {
		ms := re.FindAll(input, -1)
		for _, m := range ms {
			matchid++
			t.Logger.Info("Match #%d: %s", matchid, string(m))
		}
	}
}

func (t *TcpProxy) createReplacer(replace string) func([]byte) []byte {
	if replace == "" {
		return nil
	}
	//split by / (TODO: allow slash escapes)
	parts := strings.Split(replace, "~")
	if len(parts) != 2 {
		t.Logger.Warn("Invalid replace option")
		return nil
	}

	re, err := regexp.Compile(string(parts[0]))
	if err != nil {
		t.Logger.Warn("Invalid replace regex: %s", err)
		return nil
	}

	repl := []byte(parts[1])

	t.Logger.Info("Replacing %s with %s", re.String(), repl)
	return func(input []byte) []byte {
		return re.ReplaceAll(input, repl)
	}
}
