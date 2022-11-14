package api_schema

import (
	"time"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/akita-libs/akinet"
	"github.com/akitasoftware/akita-libs/client_telemetry"
	"github.com/akitasoftware/akita-libs/spec_summary"
	"github.com/akitasoftware/akita-libs/tags"
	"github.com/akitasoftware/akita-libs/time_span"
)

// NetworkDirection is always relative to subject service.
type NetworkDirection string

const (
	Inbound NetworkDirection = "INBOUND"
)

type APISpecState string

const (
	APISpecInitialized APISpecState = "INITIALIZED"
	APISpecComputing   APISpecState = "COMPUTING"
	APISpecDone        APISpecState = "DONE"
	APISpecError       APISpecState = "ERROR"
)

type LearnMode string

const (
	LearnedLearnMode          LearnMode = "LEARNED"
	LearnedWithEditsLearnMode LearnMode = "LEARNED_WITH_EDITS"
	MergedLearnMode           LearnMode = "MERGED"
)

// References an API spec by ID or version. Only one field may be set.
type APISpecReference struct {
	ID      *akid.APISpecID `json:"id,omitempty"`
	Version *string         `json:"version,omitempty"`
}

// Also used as a model in specs_db.
type APISpecVersion struct {
	//lint:ignore U1000 Used by pg-go
	tableName struct{} `pg:"api_spec_versions" json:"-"`

	Name         string         `pg:"name" json:"name"`
	APISpecID    akid.APISpecID `pg:"api_spec_id" json:"api_spec_id"`
	ServiceID    akid.ServiceID `pg:"service_id,type:uuid" json:"service_id"`
	CreationTime time.Time      `pg:"creation_time" json:"creation_time"`
}

type CheckpointRequest struct {
	// Optional: name to assign to the API spec generated by the checkpoint.
	APISpecName string `json:"api_spec_name"`
}

type CheckpointResponse struct {
	APISpecID akid.APISpecID `json:"api_spec_id"`
}

type CreateLearnSessionRequest struct {
	// Optional argument that specifies an existing API spec that specs generated
	// from this learn session should extend upon.
	BaseAPISpecRef *APISpecReference `json:"base_api_spec_ref,omitempty"`

	// Optional key-value pairs to tag this learn session.
	// We reserve tags with "x-akita-" prefix for internal use.
	Tags tags.SingletonTags `json:"tags,omitempty"`

	// Optional name for the learn session.
	Name string `json:"name"`
}

type LearnSession struct {
	ID           akid.LearnSessionID `json:"id"`
	Name         string              `json:"name"`
	IdentityID   akid.IdentityID     `json:"identity_id"`
	ServiceID    akid.ServiceID      `json:"service_id"`
	CreationTime time.Time           `json:"creation_time"`

	// Optional field whose presence indicates that the learn session is an
	// extension to an existing API spec.
	BaseAPISpecID *akid.APISpecID `json:"base_api_spec_id,omitempty"`

	// HasMany relationship.
	Tags []LearnSessionTag `json:"tags"`
}

func NewLearnSession(
	ID akid.LearnSessionID,
	Name string,
	IdentityID akid.IdentityID,
	ServiceID akid.ServiceID,
	CreationTime time.Time,

	BaseAPISpecID *akid.APISpecID,

	Tags []LearnSessionTag,
) *LearnSession {
	return &LearnSession{
		ID:           ID,
		Name:         Name,
		IdentityID:   IdentityID,
		ServiceID:    ServiceID,
		CreationTime: CreationTime,

		BaseAPISpecID: BaseAPISpecID,

		Tags: Tags,
	}
}

// An extended version of LearnSession that includes extra information for
// presentation in the web console when listing learn sessions.
// XXX Should inherit from LearnSession, but Go. :(
type ListedLearnSession struct {
	ID           akid.LearnSessionID `json:"id"`
	Name         string              `json:"name"`
	IdentityID   akid.IdentityID     `json:"identity_id"`
	ServiceID    akid.ServiceID      `json:"service_id"`
	CreationTime time.Time           `json:"creation_time"`

	// Optional field whose presence indicates that the learn session is an
	// extension to an existing API spec.
	BaseAPISpecID *akid.APISpecID `json:"base_api_spec_id,omitempty"`

	// HasMany relationship.
	Tags []LearnSessionTag `json:"tags"`

	// Identifies the set of API specs that are derived from this learn session.
	APISpecs []akid.APISpecID `json:"api_spec_ids"`

	Stats *LearnSessionStats `json:"stats,omitempty"`
}

func NewListedLearnSession(
	ID akid.LearnSessionID,
	Name string,
	IdentityID akid.IdentityID,
	ServiceID akid.ServiceID,
	CreationTime time.Time,

	BaseAPISpecID *akid.APISpecID,

	Tags []LearnSessionTag,
	APISpecs []akid.APISpecID,
	Stats *LearnSessionStats,
) *ListedLearnSession {
	return &ListedLearnSession{
		ID:           ID,
		Name:         Name,
		IdentityID:   IdentityID,
		ServiceID:    ServiceID,
		CreationTime: CreationTime,

		BaseAPISpecID: BaseAPISpecID,

		Tags:     Tags,
		APISpecs: APISpecs,
		Stats:    Stats,
	}
}

// Statistics about a single learn session.
type LearnSessionStats struct {
	// The number of witnesses in this learn session.
	NumWitnesses int `json:"num_witnesses"`
}

func NewLearnSessionStats(NumWitnesses int) *LearnSessionStats {
	return &LearnSessionStats{
		NumWitnesses: NumWitnesses,
	}
}

type LearnSessionTag struct {
	//lint:ignore U1000 Used by pg-go
	tableName struct{} `pg:"learn_session_tags"`

	LearnSessionID akid.LearnSessionID `pg:"learn_session_id,type:uuid" json:"learn_session_id"`
	Key            tags.Key            `pg:"key" json:"key"`
	Value          string              `pg:"value,use_zero" json:"value"`
}

type CreateSpecRequest struct {
	// Learn sessions to create spec from.
	LearnSessionIDs []akid.LearnSessionID `json:"learn_session_ids"`

	// Optional: name to assign to the API spec generated by the checkpoint.
	Name string `json:"name"`

	// Optional: user-specified tags.
	Tags tags.SingletonTags `json:"tags"`
}

type CreateTimeSpanSpecRequest struct {
	Deployment string                     `json:"deployment"`
	TimeSpan   time_span.HalfOpenInterval `json:"time_span"`
}

type UploadSpecRequest struct {
	Name string             `json:"name"`
	Tags tags.SingletonTags `json:"tags,omitempty"`

	// TODO(kku): use multipart/form-data upload once we can support it.
	Content string `json:"content"`
}

type UploadSpecResponse struct {
	ID akid.APISpecID `json:"id"`
}

type ListSessionsResponse struct {
	Sessions []*ListedLearnSession `json:"sessions"`
}

type CreateSpecResponse struct {
	ID akid.APISpecID `json:"id"`
}

type GetSpecMetadataResponse struct {
	Name string `json:"name"`

	State APISpecState `json:"state"`

	// Deprecated in favor of TagsSet, which supports multiple values
	// per tag.
	Tags    tags.SingletonTags `json:"tags,omitempty"`
	TagsSet tags.Tags          `json:"tags_set,omitempty"`
}

type GetSpecResponse struct {
	Content string `json:"content"`

	// TODO: remove
	// If the spec was created from a learn session, the session's ID is included.
	LearnSessionID *akid.LearnSessionID `json:"learn_session_id,omitempty"`

	// If the spec was created from a learn session, the session's ID is included.
	// If the spec was created by merging other API specs, those spec's session
	// IDs are included.
	LearnSessionIDs []akid.LearnSessionID `json:"learn_session_ids,omitempty"`

	Name string `json:"name"`

	State APISpecState `json:"state"`

	Summary *spec_summary.Summary `json:"summary,omitempty"`

	// Deprecated in favor of TagsSet, which supports multiple values
	// per tag.
	Tags    tags.SingletonTags `json:"tags,omitempty"`
	TagsSet tags.Tags          `json:"tags_set,omitempty"`
}

type SetSpecVersionRequest struct {
	APISpecID akid.APISpecID `json:"api_spec_id"`
}

// Types used in spec list
type SpecInfo struct {
	ID akid.APISpecID `json:"id"`

	Name string `json:"name,omitempty"`

	// Deprecated in favor of TagsSet, which supports multiple values
	// per tag.
	Tags    tags.SingletonTags `json:"tags,omitempty"`
	TagsSet tags.Tags          `json:"tags_set,omitempty"`

	VersionTags []string `json:"version_tags,omitempty"`

	CreationTime time.Time    `json:"creation_time"`
	EditTime     time.Time    `json:"edit_time"`
	State        APISpecState `json:"state"`

	// Model and trace statistics
	// Number of endpoints.
	NumEndpoints int `json:"num_endpoints,omitempty"`

	// Deployment times
	TraceStartTime *time.Time `json:"trace_start_time,omitempty"`
	TraceEndTime   *time.Time `json:"trace_end_time,omitempty"`
}

type ListSpecsResponse struct {
	Specs []SpecInfo `json:"specs"`

	// Indicates whether this is the last page of results. Always true when the
	// results are not paginated.
	LastPage bool `json:"last_page"`
}

type TimelineAggregation string
type TimelineValue string

// Time along with a map of values such as
// count, latency, etc.
type TimelineEvent struct {
	Time   time.Time      `json:"time"`
	Values TimelineValues `json:"values"`
}

// Tracks summary statistics for a group of events, as derived from an
// EndpointStats object. Each statistic is optionally populated.
type TimelineValues struct {
	NumEvents       *int     `json:"count,omitempty"`
	EventsPerMinute *float32 `json:"rate,omitempty"`

	MaxLatency    *float32 `json:"latency_max,omitempty"`
	MinLatency    *float32 `json:"latency_min,omitempty"`
	MeanLatency   *float32 `json:"latency_mean,omitempty"`
	MedianLatency *float32 `json:"latency_median,omitempty"`
	P90Latency    *float32 `json:"latency_90p,omitempty"`
	P95Latency    *float32 `json:"latency_95p,omitempty"`
	P99Latency    *float32 `json:"latency_99p,omitempty"`

	MaxRTT    *float32 `json:"rtt_max,omitempty"`
	MinRTT    *float32 `json:"rtt_min,omitempty"`
	MeanRTT   *float32 `json:"rtt_mean,omitempty"`
	MedianRTT *float32 `json:"rtt_median,omitempty"`
	P90RTT    *float32 `json:"rtt_90p,omitempty"`
	P95RTT    *float32 `json:"rtt_95p,omitempty"`
	P99RTT    *float32 `json:"rtt_99p,omitempty"`

	// The number of calls that resulted in a 4XX response.
	Num4xx *int `json:"num_4xx,omitempty"`

	// The number of calls that resulted in a 5XX response.
	Num5xx *int `json:"num_5xx,omitempty"`
}

// These arguments may be given as the "aggregate" query parameter.
// They correspond with the keys in the response below.
const (
	Aggr_Count     TimelineAggregation = "count"     // count of events within bucket
	Aggr_Count_4XX TimelineAggregation = "count_4xx" // count of HTTP events with a 4XX status code
	Aggr_Count_5XX TimelineAggregation = "count_5xx" // count of HTTP events with a 5XX status code
	Aggr_Rate      TimelineAggregation = "rate"      // rate in events per minute
	Aggr_Max       TimelineAggregation = "max"       // max of latency and RTT
	Aggr_Min       TimelineAggregation = "min"       // min of latency and RTT
	Aggr_Mean      TimelineAggregation = "mean"      // arithmetic mean of latency and RTT
	Aggr_Median    TimelineAggregation = "median"    // median value of latency and RTT
	Aggr_90p       TimelineAggregation = "90p"       // 90th percentile latency and RTT
	Aggr_95p       TimelineAggregation = "95p"       // 95th percentile latency and RTT
	Aggr_99p       TimelineAggregation = "99p"       // 99th percentile latency and RTT
)

// These are the available keys for Timeline.Values.
const (
	Event_Count                TimelineValue = "count"                // count of events within bucket
	Event_Rate                 TimelineValue = "rate"                 // rate in events per minute
	Event_Latency              TimelineValue = "latency"              // processing latency in milliseconds
	Event_Latency_Max          TimelineValue = "latency_max"          // maximum latency
	Event_Latency_Min          TimelineValue = "latency_min"          // minimum latency
	Event_Latency_Mean         TimelineValue = "latency_mean"         // arithmetic mean latency
	Event_Latency_Median       TimelineValue = "latency_median"       // median (50th percentile) latency
	Event_Latency_90p          TimelineValue = "latency_90p"          // 90th percentile latency
	Event_Latency_95p          TimelineValue = "latency_95p"          // 95th percentile latency
	Event_Latency_99p          TimelineValue = "latency_99p"          // 99th percentile latency
	Event_RTT                  TimelineValue = "rtt"                  // estimated network round-trip time, in milliseconds
	Event_RTT_Max              TimelineValue = "rtt_max"              // maximum rtt
	Event_RTT_Min              TimelineValue = "rtt_min"              // minimum rtt
	Event_RTT_Mean             TimelineValue = "rtt_mean"             // arithmetic mean rtt
	Event_RTT_Median           TimelineValue = "rtt_median"           // median (50th percentile) rtt
	Event_RTT_90p              TimelineValue = "rtt_90p"              // 90th percentile rtt
	Event_RTT_95p              TimelineValue = "rtt_95p"              // 95th percentile rtt
	Event_RTT_99p              TimelineValue = "rtt_99p"              // 99th percentile rtt
	Event_Latency_1ms_Count    TimelineValue = "latency_1ms_count"    // Count of events with latency <=1ms
	Event_Latency_10ms_Count   TimelineValue = "latency_10ms_count"   // Count of events with latency >1ms and <=10ms
	Event_Latency_100ms_Count  TimelineValue = "latency_100ms_count"  // Count of events with latency >10ms and <=100ms
	Event_Latency_1000ms_Count TimelineValue = "latency_1000ms_count" // Count of events with latency >100ms and <=1s
	Event_Latency_Inf_Count    TimelineValue = "latency_inf_count"    // Count of events with latency >1 second

)

// Describes a common set of attributes for a group of endpoints. If a
// non-default value (i.e., non-empty string or non-zero int) is provided for
// any attribute in this struct, then all endpoints in the group will have that
// value for that attribute.
//
// XXX Would be nice for this to just be map[attribute]string, but then we
// wouldn't be able to use this as a map key.
type EndpointGroupAttributes struct {
	Method       string `json:"method,omitempty"`
	Host         string `json:"host,omitempty"`
	PathTemplate string `json:"path_template,omitempty"`
	ResponseCode int    `json:"response_code,omitempty"`
}

type Timeline struct {
	// Describes the common set of attributes for the endpoints in this timeline.
	GroupAttributes EndpointGroupAttributes `json:"group_attrs"`

	// Events in time order
	Events []TimelineEvent `json:"events"`
}

type TimelineResponse struct {
	// Report what time range is included in this responses; less than
	// request if data is not available or limit was hit.
	ActualStartTime time.Time `json:"actual_start_time"`
	ActualEndTime   time.Time `json:"actual_end_time"`

	// One timeline per selected endpoint
	Timelines []Timeline `json:"timelines"`

	// If incomplete due to limit, the first unreported start time
	NextStartTime *time.Time `json:"next_start_time,omitempty"`

	// Time spans that have incomplete data.
	IncompleteSpans []time_span.HalfOpenInterval `json:"incomplete_spans,omitempty"`
}

// An HTTP request and response between two nodes in a graph.
type HTTPGraphEdge struct {
	// Describe the source and destination vertices by the attributes
	// they share in common: either just Host for a service-level vertex,
	// or Host + Method + PathTemplate for a end-point level vertex, although
	// we could imagine other possibilities.
	SourceAttributes EndpointGroupAttributes `json:"source_attrs"`
	TargetAttributes EndpointGroupAttributes `json:"target_attrs"`

	// Aggregate values attached to the edge, e.g., "count"
	Values map[TimelineValue]float32 `json:"values"`
}

// Represents a TCP connection between two nodes in a graph.
type TCPGraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`

	// If true, the source is known to have initiated the connection. Otherwise,
	// the "source" and "target" designations are chosen so that `source` <=
	// `target`. One way to render this is to use a directed edge if
	// "InitiatorKnown" is true, and an undirected edge if false.
	InitiatorKnown bool `json:"initiator_known"`

	// Aggregate values attached to the edge, e.g., "count"
	Values map[TimelineValue]float32 `json:"values"`
}

// Represents a TLS connection between two nodes in a graph.
type TLSGraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`

	TLSVersion                    akinet.TLSVersion `json:"tls_version"`
	NegotiatedApplicationProtocol *string           `json:"negotiated_application_protocol"`

	// Aggregate values attached to the edge, e.g., "count"
	Values map[TimelineValue]float32 `json:"values"`
}

type GraphResponse struct {
	// Graph edges representing HTTP requests and responses.
	HTTPEdges []HTTPGraphEdge `json:"edges"`

	// Graph edges representing TCP connections.
	TCPEdges []TCPGraphEdge `json:"tcp_edges"`

	// Graph edges representing TLS connections.
	TLSEdges []TLSGraphEdge `json:"tls_edges"`

	// TODO: vertex list? vertex or edge count?
	// TODO: pagination
}

func (g *GraphResponse) NumEdges() int {
	return len(g.HTTPEdges) + len(g.TCPEdges)
}

func (g *GraphResponse) IsEmpty() bool {
	return g.NumEdges() == 0
}

type ApidumpErrorType string

const (
	// We'd like to report on "service not found", or permissions errors, but
	// without a working service ID we can't access the telemetry endpoint.
	// We'll have to rely on Segment telemetry for those.

	ApidumpError_PCAPPermission ApidumpErrorType = "PCAP permission failure" // Cannot obtain permission for packet capture

	// Cannot access network interface due to "Function not implemented" error.
	// Often indicates the agent was built for a different architecture and is
	// running in a partially emulated environment.
	ApidumpError_PCAPInterfaceNotImplemented ApidumpErrorType = "PCAP interface not implemented"

	ApidumpError_PCAPInterfaceOther ApidumpErrorType = "Other PCAP interface failure" // Cannot access interface for packet capture
	ApidumpError_InvalidFilters     ApidumpErrorType = "Invalid filters"              // Error parsing filters
	ApidumpError_TraceCreation      ApidumpErrorType = "Trace creation failure"       // Can't create learn session
	ApidumpError_Other              ApidumpErrorType = "Other error"
)

type PostClientPacketCaptureStatsRequest struct {
	ClientID                  akid.ClientID `json:"client_id"`
	ObservedStartingAt        time.Time     `json:"observed_starting_at"`
	ObservedDurationInSeconds int           `json:"observed_duration_in_seconds"`

	// If PacketCountSummary is absent, then this observation period has
	// started but not yet concluded.
	PacketCountSummary *client_telemetry.PacketCountSummary `json:"packet_count_summary,omitempty"`

	// Report on any errors encounted during apidump that should be shown to the user.
	ApidumpError     ApidumpErrorType `json:"apidump_error,omitempty"`
	ApidumpErrorText string           `json:"apidump_error_text,omitempty"`
}

type PostInitialClientTelemetry struct {
	ClientID                  akid.ClientID `json:"client_id"`
	ObservedStartingAt        time.Time     `json:"observed_starting_at"`
	ObservedDurationInSeconds int           `json:"observed_duration_in_seconds"`

	CLIVersion    string `json:"cli_version"`
	CLITargetArch string `json:"cli_target_arch"`

	// True when the CLI is running in a Docker container we distributed, e.g.
	// akitasoftware/cli:latest.
	AkitaDockerRelease bool `json:"akita_docker_release"`

	// True when the host.docker.internal DNS name resolves, which indicates the
	// CLI is running in a docker container hosted on Docker Desktop.
	DockerDesktop bool `json:"docker_desktop"`
}

type UserResponse struct {
	ID             akid.UserID         `json:"id"`
	OrganizationID akid.OrganizationID `json:"organization_id"`
	Name           string              `json:"name"`
	Email          string              `json:"email"`
	CreatedAt      time.Time           `json:"created_at"`
	TOSAcceptedAt  time.Time           `json:"tos_accepted_at"`
}
