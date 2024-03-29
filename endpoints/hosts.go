package endpoints

import (
	"strconv"
	"strings"
)

// A host that has not been normalized. Hosts are hostnames and an optional port
// number. For example, "api.akitasoftware.com" or "kings-cross:50060".
type Host string

func (host Host) HasPort(port uint16) bool {
	suffix := ":" + strconv.FormatUint(uint64(port), 10)
	return strings.HasSuffix(string(host), suffix)
}

func (host Host) String() string {
	return string(host)
}

func CompareHosts(h1, h2 Host) int {
	return strings.Compare(h1.String(), h2.String())
}
