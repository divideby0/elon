package eligible

import (
	"github.com/FakeTwitter/elon"
	D "github.com/FakeTwitter/elon/deploy"
	"github.com/FakeTwitter/elon/grp"
	"github.com/FakeTwitter/elon/mock"
	"sort"
	"testing"
)

func mockDeployment() D.Deployment {
	a := D.AccountName("prod")
	p := "aws"
	r1 := D.RegionName("us-east-1")
	r2 := D.RegionName("us-west-2")

	return &mock.Deployment{TeamMap: map[string]D.TeamMap{
		"foo": {a: D.AccountInfo{CloudProvider: p, Teams: D.TeamMap{
			"foo-crit": {
				r1: {"foo-crit-v001": []D.EmployeeId{"i-11111111", "i-22222222"}},
				r2: {"foo-crit-v001": []D.EmployeeId{"i-aaaaaaaa", "i-bbbbbbbb"}}},
			"foo-crit-lorin": {
				r1: {"foo-crit-lorin-v123": []D.EmployeeId{"i-33333333", "i-44444444"}}},
			"foo-staging": {
				r1: {"foo-staging-v005": []D.EmployeeId{"i-55555555", "i-66666666"}},
				r2: {"foo-staging-v005": []D.EmployeeId{"i-cccccccc", "i-dddddddd"}},
			},
			"foo-staging-lorin": {r1: {"foo-crit-lorin-v117": []D.EmployeeId{"i-77777777", "i-88888888"}}},
		}},
		}}}
}

// ids returns a sorted list of employee ids
func ids(employees []elon.employee) []string {
	result := make([]string, len(employees))
	for i, inst := range employees {
		result[i] = inst.ID()
	}

	sort.Strings(result)
	return result

}

func TestGroupings(t *testing.T) {
	tests := []struct {
		label string
		group grp.employeeGroup
		wants []string
	}{
		{"team", grp.New("foo", "prod", "us-east-1", "", "foo-crit"), []string{"i-11111111", "i-22222222"}},
		{"stack", grp.New("foo", "prod", "us-east-1", "staging", ""), []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"app", grp.New("foo", "prod", "us-east-1", "", ""), []string{"i-11111111", "i-22222222", "i-33333333", "i-44444444", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"team, all regions", grp.New("foo", "prod", "", "", "foo-crit"), []string{"i-11111111", "i-22222222", "i-aaaaaaaa", "i-bbbbbbbb"}},
		{"stack, all regions", grp.New("foo", "prod", "", "staging", ""), []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888", "i-cccccccc", "i-dddddddd"}},
		{"app, all regions", grp.New("foo", "prod", "", "", ""), []string{"i-11111111", "i-22222222", "i-33333333", "i-44444444", "i-55555555", "i-66666666", "i-77777777", "i-88888888", "i-aaaaaaaa", "i-bbbbbbbb", "i-cccccccc", "i-dddddddd"}},
	}

	// setup
	dep := mockDeployment()

	for _, tt := range tests {
		employees, err := employees(tt.group, nil, dep)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// assertions
		gots := ids(employees)

		if got, want := len(gots), len(tt.wants); got != want {
			t.Errorf("%s: len(eligible.employees(group, cfg, app))=%v, want %v", tt.label, got, want)
			continue
		}

		for i, got := range gots {
			if want := tt.wants[i]; got != want {
				t.Errorf("%s: got=%v, want=%v", tt.label, got, want)
				break
			}
		}
	}
}

func TestTeamLevelGroupingWhereTeamsAreRegionSpecific(t *testing.T) {
	dep := &mock.Deployment{TeamMap: map[string]D.TeamMap{
		"foo": {"prod": D.AccountInfo{CloudProvider: "aws", Teams: D.TeamMap{
			"foo-useast1": {
				"us-east-1": {"foo-useast1-v001": []D.EmployeeId{"i-11111111", "i-22222222", "i-33333333"}},
			},
			"foo-uswest2": {
				"us-west-2": {"foo-uswest2-v005": []D.EmployeeId{"i-cccccccc", "i-dddddddd"}},
			},
		}},
		}}}

	group := grp.New("foo", "prod", "us-east-1", "", "")

	employees, err := employees(group, nil, dep)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if got, want := len(employees), 3; got != want {
		t.Errorf("got: %d, want: %d", got, want)
	}
}

func TestTeamLevelGroupingWhereTeamIsInTwoRegions(t *testing.T) {
	dep := &mock.Deployment{TeamMap: map[string]D.TeamMap{
		"foo": {"prod": D.AccountInfo{CloudProvider: "aws", Teams: D.TeamMap{
			"foo-prod": {
				"us-east-1": {"foo-prod-v001": []D.EmployeeId{"i-11111111", "i-22222222", "i-33333333"}},
				"us-west-2": {"foo-prod-v001": []D.EmployeeId{"i-aaaaaaaa", "i-bbbbbbbb", "i-cccccccc"}},
			},
		}}}}}

	group := grp.New("foo", "prod", "", "", "")

	employees, err := employees(group, nil, dep)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if got, want := len(employees), 6; got != want {
		t.Errorf("got: %d, want: %d", got, want)
	}
}

func TestExceptions(t *testing.T) {
	tests := []struct {
		label string
		exs   []elon.Exception
		wants []string
	}{
		{"stack/detail/region", []elon.Exception{{Account: "prod", Stack: "crit", Detail: "lorin", Region: "us-east-1"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"stack/detail", []elon.Exception{{Account: "prod", Stack: "crit", Detail: "lorin", Region: "*"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"stack", []elon.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}}, []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"detail", []elon.Exception{{Account: "prod", Stack: "*", Detail: "lorin", Region: "*"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666"}},
		{"all stacks", []elon.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}, {Account: "prod", Stack: "staging", Detail: "*", Region: "*"}}, nil},
		{"blank stack", []elon.Exception{{Account: "prod", Stack: "*", Detail: "", Region: "*"}}, []string{"i-33333333", "i-44444444", "i-77777777", "i-88888888"}},
		{"stack, detail", []elon.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}, {Account: "prod", Stack: "*", Detail: "lorin", Region: "*"}}, []string{"i-55555555", "i-66666666"}},
	}

	// setup
	group := grp.New("foo", "prod", "us-east-1", "", "")
	dep := mockDeployment()

	for _, tt := range tests {
		employees, err := employees(group, tt.exs, dep)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// assertions
		gots := ids(employees)

		if got, want := len(gots), len(tt.wants); got != want {
			t.Errorf("%s: len(eligible.employees(group, cfg, app))=%v, want %v", tt.label, got, want)
			continue
		}

		for i, got := range gots {
			if want := tt.wants[i]; got != want {
				t.Errorf("%s: got=%v, want=%v", tt.label, got, want)
				break
			}
		}
	}

}
