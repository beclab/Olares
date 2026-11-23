package access

import "testing"

func TestValidateHTTPRules(t *testing.T) {
	if err := ValidateHTTPRules(nil); err == nil {
		t.Fatal("expected empty rules error")
	}
	ok := []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypeExact}}
	if err := ValidateHTTPRules(ok); err != nil {
		t.Fatalf("exact: %v", err)
	}
	if err := ValidateHTTPRules([]HTTPRule{{Method: "GET", Path: "/api/", PathType: PathTypePrefix}}); err == nil {
		t.Fatal("expected trailing slash prefix rejection")
	}
	if err := ValidateHTTPRules([]HTTPRule{{Method: "GET", Path: "/x", PathType: PathTypeRegularExpression}}); err == nil {
		t.Fatal("expected regex prefix ^/ rejection")
	}
	if err := ValidateHTTPRules([]HTTPRule{{Method: "GET", Path: "^/state$", PathType: PathTypeRegularExpression}}); err != nil {
		t.Fatalf("regex: %v", err)
	}
	tooMany := make([]HTTPRule, MaxRulesPerCapability+1)
	for i := range tooMany {
		tooMany[i] = HTTPRule{Method: "GET", Path: "/x", PathType: PathTypeExact}
	}
	if err := ValidateHTTPRules(tooMany); err == nil {
		t.Fatal("expected capacity error")
	}
}

func TestValidateManagedObjectBudget(t *testing.T) {
	if err := ValidateManagedObjectBudget(MaxManagedObjectsPerPolicy); err != nil {
		t.Fatal(err)
	}
	if err := ValidateManagedObjectBudget(MaxManagedObjectsPerPolicy + 1); err == nil {
		t.Fatal("expected budget error")
	}
}

func TestPolicyCompilationEnabled(t *testing.T) {
	if PolicyCompilationEnabled("false") || PolicyCompilationEnabled("") {
		t.Fatal("default must be disabled")
	}
	if !PolicyCompilationEnabled("true") {
		t.Fatal("true must enable")
	}
}
