package spec_summary

import (
	"testing"

	"github.com/akitasoftware/akita-libs/akid"
	"github.com/akitasoftware/go-utils/optionals"
	"github.com/stretchr/testify/assert"
)

func TestFiltersToMethods(t *testing.T) {
	fm := NewFiltersToMethods[string]()

	m1 := akid.GenerateAPIMethodID().GetUUID().String()
	m2 := akid.GenerateAPIMethodID().GetUUID().String()

	fm.InsertNondirectionalFilter(HostFilter, "example.com", optionals.Some(m1))
	fm.InsertNondirectionalFilter(HostFilter, "example.com", optionals.Some(m2))
	fm.InsertNondirectionalFilter(HostFilter, "example.org", optionals.None[string]())
	fm.InsertDirectionalFilter(RequestDirection, AuthFilter, "None", optionals.Some(m1))
	fm.InsertDirectionalFilter(RequestDirection, AuthFilter, "None", optionals.Some(m2))
	fm.InsertDirectionalFilter(RequestDirection, AuthFilter, "Basic", optionals.None[string]())

	fm.filterMap.Insert(HttpMethodFilter, "GET", optionals.Some(m1))
	fm.filterMap.Insert(HttpMethodFilter, "POST", optionals.Some(m2))

	// No filters.
	directedSummary, numMethods := fm.SummarizeWithFilters(nil)

	assert.Equal(t, 2, numMethods)
	assert.Equal(t, 2, directedSummary.NondirectedFilters[HostFilter]["example.com"], "directed: example")
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HttpMethodFilter]["GET"], "directed: get")
	assert.Equal(t, 2, directedSummary.DirectedFilters[RequestDirection][AuthFilter]["None"], "directed: auth")

	exampleOrgCount, exampleOrgExists := directedSummary.NondirectedFilters[HostFilter]["example.org"]
	assert.Equal(t, 0, exampleOrgCount, "directed: example.org")
	assert.True(t, exampleOrgExists, "directed: example.org")

	_, exampleNetExists := directedSummary.NondirectedFilters[HostFilter]["example.net"]
	assert.False(t, exampleNetExists, "directed: example.net")

	basicCount, basicExists := directedSummary.DirectedFilters[RequestDirection][AuthFilter]["Basic"]
	assert.Equal(t, 0, basicCount, "directed: basic")
	assert.True(t, basicExists, "directed: basic")

	_, bearerExists := directedSummary.DirectedFilters[RequestDirection][AuthFilter]["Bearer"]
	assert.False(t, bearerExists, "directed: bearer")

	summary := directedSummary.ToSummary()

	assert.Equal(t, 2, summary.Hosts["example.com"], "example")
	assert.Equal(t, 1, summary.HTTPMethods["GET"], "get")

	// With filters.
	directedSummary, numMethods = fm.SummarizeWithFilters(Filters{
		HttpMethodFilter: {"GET"},
	})

	assert.Equal(t, 1, numMethods)
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HostFilter]["example.com"], "directed: example")
	assert.Equal(t, 1, directedSummary.NondirectedFilters[HttpMethodFilter]["GET"], "directed: get")
	assert.Equal(t, 0, directedSummary.NondirectedFilters[HttpMethodFilter]["PUT"], "directed: put")
	assert.Equal(t, 1, directedSummary.DirectedFilters[RequestDirection][AuthFilter]["None"], "directed: auth")

	summary = directedSummary.ToSummary()

	assert.Equal(t, 1, summary.Hosts["example.com"], "example")
	assert.Equal(t, 1, summary.HTTPMethods["GET"], "get")
	assert.Equal(t, 0, summary.HTTPMethods["PUT"], "put")
}
