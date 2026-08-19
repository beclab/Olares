package router

import (
	"bytes"
	"strings"
	"testing"
)

func TestSeveralGroupingsAreAcceptedInTheOrderAsked(t *testing.T) {
	dims, err := parseSpendDims("model, provider ,user")
	if err != nil {
		t.Fatalf("parseSpendDims: %v", err)
	}
	want := []string{"model", "provider", "user"}
	if len(dims) != len(want) {
		t.Fatalf("got %v, want %v", dims, want)
	}
	for i := range want {
		if dims[i] != want[i] {
			t.Fatalf("got %v, want %v", dims, want)
		}
	}
}

func TestARepeatedGroupingIsNotReportedTwice(t *testing.T) {
	dims, err := parseSpendDims("model,MODEL,model")
	if err != nil {
		t.Fatalf("parseSpendDims: %v", err)
	}
	if len(dims) != 1 || dims[0] != "model" {
		t.Fatalf("got %v, want one model", dims)
	}
}

// Router refuses hour alongside anything else, so the CLI has to say why here
// rather than forward a request it knows will come back 400.
func TestHourCannotShareARequest(t *testing.T) {
	if _, err := parseSpendDims("day,hour"); err == nil {
		t.Fatal("expected hour to be refused alongside day")
	} else if !strings.Contains(err.Error(), "on its own") {
		t.Fatalf("error should say hour is answered alone, got: %v", err)
	}
	if _, err := parseSpendDims("hour"); err != nil {
		t.Fatalf("hour alone must still work: %v", err)
	}
}

func TestAnUnknownGroupingNamesTheValidOnes(t *testing.T) {
	_, err := parseSpendDims("model,tag")
	if err == nil {
		t.Fatal("expected tag to be refused")
	}
	for _, dim := range spendDims {
		if !strings.Contains(err.Error(), dim) {
			t.Fatalf("error should list %q, got: %v", dim, err)
		}
	}
}

func TestAnEmptyGroupingIsRefused(t *testing.T) {
	if _, err := parseSpendDims(" , "); err == nil {
		t.Fatal("expected an empty --by to be refused")
	}
}

// The totals belong to the whole answer, not to any one table, so they are
// printed once however many groupings were asked for.
func TestTheTotalsAreStatedOnceAcrossGroupings(t *testing.T) {
	multi := &spendMultiSummary{
		TotalRequests:        10,
		TotalSuccessRequests: 8,
		TotalCostUSD:         1.5,
		TotalTokens:          400,
	}
	multi.Dims = map[string]struct {
		Items     []spendSummaryRow `json:"items"`
		Truncated bool              `json:"truncated"`
	}{
		"model":    {Items: []spendSummaryRow{{Key: "m", Label: "Model", Requests: 6}}},
		"provider": {Items: []spendSummaryRow{{Key: "p", Label: "Provider", Requests: 10}}},
	}
	var buf bytes.Buffer
	if err := renderUsageSummaries(&buf, []string{"model", "provider"}, multi); err != nil {
		t.Fatalf("renderUsageSummaries: %v", err)
	}
	out := buf.String()
	if got := strings.Count(out, "10 requests, 2 of them failed"); got != 1 {
		t.Fatalf("totals should appear once, appeared %d times in:\n%s", got, out)
	}
	for _, header := range []string{"MODEL", "PROVIDER"} {
		if !strings.Contains(out, header) {
			t.Fatalf("missing the %s table in:\n%s", header, out)
		}
	}
}

// A grouping that was asked for and not answered is said out loud: an empty
// table would read as "no calls were grouped this way".
func TestAGroupingRouterDidNotAnswerIsNamed(t *testing.T) {
	multi := &spendMultiSummary{TotalRequests: 3, TotalSuccessRequests: 3}
	multi.Dims = map[string]struct {
		Items     []spendSummaryRow `json:"items"`
		Truncated bool              `json:"truncated"`
	}{
		"model": {Items: []spendSummaryRow{{Key: "m", Requests: 3}}},
	}
	var buf bytes.Buffer
	if err := renderUsageSummaries(&buf, []string{"model", "user"}, multi); err != nil {
		t.Fatalf("renderUsageSummaries: %v", err)
	}
	if !strings.Contains(buf.String(), "did not report this grouping") {
		t.Fatalf("expected the missing grouping to be named, got:\n%s", buf.String())
	}
}
