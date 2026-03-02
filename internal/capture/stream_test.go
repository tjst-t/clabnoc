package capture

import (
	"testing"
)

func TestParseTcpdumpLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantProto string
		wantSrc   string
		wantDst   string
		wantLen   int
		wantErr  bool
	}{
		{
			name:     "TCP SYN",
			line:     "12:34:56.789012 IP 10.0.0.1.443 > 10.0.0.2.52341: Flags [S.], seq 1234, ack 5678, win 65535, length 0",
			wantProto: "TCP",
			wantSrc:   "10.0.0.1.443",
			wantDst:   "10.0.0.2.52341",
			wantLen:   0,
		},
		{
			name:     "TCP with data",
			line:     "12:34:56.789012 IP 192.168.1.1.80 > 192.168.1.2.54321: Flags [P.], seq 1:101, ack 1, win 65535, length 100",
			wantProto: "TCP",
			wantSrc:   "192.168.1.1.80",
			wantDst:   "192.168.1.2.54321",
			wantLen:   100,
		},
		{
			name:     "UDP DNS",
			line:     "12:34:56.789012 IP 10.0.0.1.53 > 10.0.0.2.12345: UDP, length 64",
			wantProto: "UDP",
			wantSrc:   "10.0.0.1.53",
			wantDst:   "10.0.0.2.12345",
			wantLen:   64,
		},
		{
			name:     "ICMP echo request",
			line:     "12:34:56.789012 IP 10.0.0.1 > 10.0.0.2: ICMP echo request, id 1234, seq 1, length 64",
			wantProto: "ICMP",
			wantSrc:   "10.0.0.1",
			wantDst:   "10.0.0.2",
			wantLen:   64,
		},
		{
			name:     "ICMP echo reply",
			line:     "12:34:56.789012 IP 10.0.0.2 > 10.0.0.1: ICMP echo reply, id 1234, seq 1, length 64",
			wantProto: "ICMP",
			wantSrc:   "10.0.0.2",
			wantDst:   "10.0.0.1",
			wantLen:   64,
		},
		{
			name:     "ARP request",
			line:     "12:34:56.789012 ARP, Request who-has 10.0.0.1 tell 10.0.0.2, length 28",
			wantProto: "ARP",
			wantSrc:   "10.0.0.2",
			wantDst:   "10.0.0.1",
			wantLen:   28,
		},
		{
			name:     "ARP reply",
			line:     "12:34:56.789012 ARP, Reply 10.0.0.1 is-at aa:bb:cc:dd:ee:ff, length 28",
			wantProto: "ARP",
			wantSrc:   "10.0.0.1",
			wantDst:   "aa:bb:cc:dd:ee:ff",
			wantLen:   28,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
		{
			name:    "unrecognized line",
			line:    "listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes",
			wantErr: true,
		},
		{
			name:     "IPv6 TCP",
			line:     "12:34:56.789012 IP6 2001:db8::1.443 > 2001:db8::2.54321: Flags [S], seq 12345, win 65535, length 0",
			wantProto: "TCP",
			wantSrc:   "2001:db8::1.443",
			wantDst:   "2001:db8::2.54321",
			wantLen:   0,
		},
		{
			name:     "BGP TCP",
			line:     "12:34:56.789012 IP 172.20.20.2.179 > 172.20.20.3.54321: Flags [P.], seq 1:20, ack 1, win 65535, length 19",
			wantProto: "TCP",
			wantSrc:   "172.20.20.2.179",
			wantDst:   "172.20.20.3.54321",
			wantLen:   19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt, err := ParseTcpdumpLine(tt.line, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.line)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkt.Protocol != tt.wantProto {
				t.Errorf("protocol: got %q, want %q", pkt.Protocol, tt.wantProto)
			}
			if pkt.Source != tt.wantSrc {
				t.Errorf("source: got %q, want %q", pkt.Source, tt.wantSrc)
			}
			if pkt.Destination != tt.wantDst {
				t.Errorf("destination: got %q, want %q", pkt.Destination, tt.wantDst)
			}
			if pkt.Length != tt.wantLen {
				t.Errorf("length: got %d, want %d", pkt.Length, tt.wantLen)
			}
			if pkt.No != 1 {
				t.Errorf("sequence: got %d, want 1", pkt.No)
			}
			if pkt.Time == "" {
				t.Error("time should not be empty")
			}
		})
	}
}

func TestParseTcpdumpLineSequenceNumbers(t *testing.T) {
	line := "12:34:56.789012 IP 10.0.0.1 > 10.0.0.2: ICMP echo request, id 1, seq 1, length 64"

	for _, seqNo := range []int{1, 42, 1000} {
		pkt, err := ParseTcpdumpLine(line, seqNo)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pkt.No != seqNo {
			t.Errorf("expected seq %d, got %d", seqNo, pkt.No)
		}
	}
}

func TestParseTcpdumpLineInfo(t *testing.T) {
	line := "12:34:56.789012 IP 10.0.0.1.443 > 10.0.0.2.52341: Flags [S.], seq 1234, ack 5678, win 65535, length 0"
	pkt, err := ParseTcpdumpLine(line, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkt.Info == "" {
		t.Error("info should not be empty")
	}
	if pkt.Info != "Flags [S.], seq 1234, ack 5678, win 65535, length 0" {
		t.Errorf("unexpected info: %q", pkt.Info)
	}
}
