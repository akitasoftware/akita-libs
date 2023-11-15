package akinet

import (
	"github.com/google/gopacket/reassembly"

	"github.com/akitasoftware/akita-libs/memview"
)

type AcceptDecision int

const (
	Reject AcceptDecision = iota
	Accept
	NeedMoreData
)

const (
	CONNECTION_TYPE_HTTP_CLIENT   = "HTTP_CLIENT"
	CONNECTION_TYPE_HTTP_SERVER   = "HTTP_SERVER"
	CONNECTION_TYPE_TLS_CLIENT    = "TLS_CLIENT"
	CONNECTION_TYPE_TLS_SERVER    = "TLS_SERVER"
	CONNECTION_TYPE_HTTP2_PREFACE = "HTTP2_PREFACE"
)

func (d AcceptDecision) String() string {
	switch d {
	case Reject:
		return "REJECT"
	case Accept:
		return "ACCEPT"
	case NeedMoreData:
		return "NEED_MORE_DATA"
	}
	return "UNKNOWN"
}

// TCPParserFactory is responsible for creating TCPParsers and deciding whether
// a TCP flow can be parsed with the parser type created by this factory.
// Implementations must be be thread-safe.
type TCPParserFactory interface {
	Name() string

	// The caller passes a slice from the current position of the TCP flow to
	// check whether this TCP flow can be parsed by this parser.
	// If this function returns NeedMoreData, the caller should supply new data by
	// appending them to the old data and send the combined result to the Accepts
	// function again (after accounting for discardFront).
	// Implementations must return a non-negative integer for discardFront. It
	// indicates to the caller that those bytes at the start of input are not
	// useful and MUST be discarded before using the parser generated by this
	// factory in the case of ACCEPT. If the implementation returns REJECT,
	// discardFront should be the the length of input.
	//
	// If no more data is forthcoming, the caller should specify isEnd=true to let
	// the factory know. In this case, the factory implementation must return
	// either ACCEPT or REJECT.
	Accepts(input memview.MemView, isEnd bool) (decision AcceptDecision, discardFront int64)

	// Creates a new TCPParser, supplying the bidirectional ID and the TCP seq/ack
	// numbers found on the FIRST packet in the TCP flow that this parser will
	// parse. This information helps to pair up parsed results generated by
	// separate parsers operating on the two TCP flows in a TCP stream (e.g.
	// pairing up HTTP request and response on the same connection).
	CreateParser(id TCPBidiID, initialSeq, initialAck reassembly.Sequence) TCPParser
}

// TCPParser converts a TCP flow into ParsedNetworkContent. It is meant to be
// SINGLE USE, meaning that a new parser instance should be used as soon as the
// current instance generates a result, even if the flow has more data
// available. For example, if we have a flow carrying 5 HTTP responses due to
// TCP connection re-use, we need to create 5 instances of TCPParser to parse
// the flow. This allows us to simplify implementations while leaving room for
// future support of flows that carry multiple types of protocols (e.g.
// websocket upgrade from HTTP/1.1).
type TCPParser interface {
	Name() string

	ConnectionType() string

	// Caller should repeatedly supply data from the TCP flow until a non-nil
	// result or an error is returned.
	//
	// The result will be non-nil when parsing is successfully completed, and
	// unused will contain the portion of all the input ever supplied that was not
	// used to generate the result (e.g. trailing bytes after an HTTP response
	// body). Additionally, totalBytesConsumed will report the number of bytes in
	// the TCP stream that were used to produce the parsed result.
	//
	// An error will be returned when parsing fails. In this case, unused will be
	// empty, and totalBytesConsumed will report the total number of bytes ever
	// supplied to the parser.
	//
	// If no more data is forthcoming, the caller should specify isEnd=true to let
	// the parser know. In this case, the parser implementation must return a
	// non-nil result or an error.
	Parse(input memview.MemView, isEnd bool) (result ParsedNetworkContent, unused memview.MemView, totalBytesConsumed int64, err error)
}

// TCPParserSelector helps to select a TCPParserFactory from a list of
// factories.
type TCPParserFactorySelector []TCPParserFactory

// SelectFactory selects a TCPParserFactory that suitable for input. The
// possible set of return values:
//   - f=nil, decision=Reject, discardFront=len(input)
//   - no factory is suitable
//   - f=nil, decision=NeedMoreData, discardFront>=0
//   - no factory returned Accept
//   - at least one factory requested more data
//   - discardFront is the MIN of all the discardFronts returned by
//     factories requesting more data
//   - f!=nil, decision=Accept, discardFront>=0
//   - a factory has been selected
//   - caller must discard discardFront number of bytes from input before
//     feeding input to the parser generated by the factory.
func (s TCPParserFactorySelector) Select(input memview.MemView, isEnd bool) (f TCPParserFactory, decision AcceptDecision, discardFront int64) {
	discardFront = -1
	allReject := true

	for _, f := range s {
		decision, df := f.Accepts(input, isEnd)
		switch decision {
		case Accept:
			return f, Accept, df
		case NeedMoreData:
			allReject = false
			if discardFront > df || discardFront == -1 {
				discardFront = df
			}
		case Reject:
			// Do nothing.
		}
	}

	if allReject {
		return nil, Reject, input.Len()
	}
	// discardFront must be >= 0 because there is at least one NeedMoreData.
	return nil, NeedMoreData, discardFront
}
