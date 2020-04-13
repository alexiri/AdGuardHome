package util

import (
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func prepareTestDir() string {
	const dir = "./agh-test"
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func TestAutoHosts(t *testing.T) {
	ah := AutoHosts{}

	dir := prepareTestDir()
	defer func() { _ = os.RemoveAll(dir) }()

	f, _ := ioutil.TempFile(dir, "")
	defer os.Remove(f.Name())
	defer f.Close()

	_, _ = f.WriteString("  127.0.0.1   host  localhost  \n")
	_, _ = f.WriteString("  ::1   localhost  \n")

	ah.Init(f.Name())
	ah.Start()
	// wait until we parse the file
	time.Sleep(50 * time.Millisecond)

	ips := ah.Process("localhost", dns.TypeA)
	assert.True(t, ips[0].Equal(net.ParseIP("127.0.0.1")))
	ips = ah.Process("newhost", dns.TypeA)
	assert.True(t, ips == nil)

	table := ah.List()
	ips, _ = table["host"]
	assert.True(t, ips[0].String() == "127.0.0.1")

	_, _ = f.WriteString("127.0.0.2   newhost\n")
	// wait until fsnotify has triggerred and processed the file-modification event
	time.Sleep(50 * time.Millisecond)

	ips = ah.Process("newhost", dns.TypeA)
	assert.True(t, ips[0].Equal(net.ParseIP("127.0.0.2")))

	a, _ := dns.ReverseAddr("127.0.0.1")
	a = strings.TrimSuffix(a, ".")
	assert.True(t, ah.ProcessReverse(a, dns.TypePTR) == "host")
	a, _ = dns.ReverseAddr("::1")
	a = strings.TrimSuffix(a, ".")
	assert.True(t, ah.ProcessReverse(a, dns.TypePTR) == "localhost")

	ah.Close()
}

func TestIP(t *testing.T) {
	assert.True(t, dnsUnreverseAddr("1.0.0.127.in-addr.arpa").Equal(net.ParseIP("127.0.0.1").To4()))
	assert.True(t, dnsUnreverseAddr("4.3.2.1.d.c.b.a.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa").Equal(net.ParseIP("::abcd:1234")))

	assert.True(t, dnsUnreverseAddr("1.0.0.127.in-addr.arpa.") == nil)
	assert.True(t, dnsUnreverseAddr(".0.0.127.in-addr.arpa") == nil)
	assert.True(t, dnsUnreverseAddr(".3.2.1.d.c.b.a.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa") == nil)
	assert.True(t, dnsUnreverseAddr("4.3.2.1.d.c.b.a.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0..ip6.arpa") == nil)
}
