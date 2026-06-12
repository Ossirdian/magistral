package lesson

import (
	"strings"
	"testing"
	"time"
)

func TestDeleteGeneratedInWindowSQLExpandsStudentIDs(t *testing.T) {
	query, args, err := deleteGeneratedInWindowSQL(
		"school-1",
		[]string{"student-1", "student-2"},
		time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 24, 18, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(query, "student_id IN (?)") {
		t.Fatalf("student ids were not expanded: %s", query)
	}
	if !strings.Contains(query, "student_id IN ($3,$4)") {
		t.Fatalf("expected expanded student id placeholders, got: %s", query)
	}
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %#v", len(args), args)
	}
	if args[2] != "student-1" || args[3] != "student-2" {
		t.Fatalf("unexpected student args: %#v", args)
	}
}

func TestDeleteGeneratedInWindowByBranchesSQLScopesBranches(t *testing.T) {
	query, args, err := deleteGeneratedInWindowByBranchesSQL(
		"school-1",
		[]string{"branch-1", "branch-2"},
		true,
		time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 24, 18, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(query, "branch_id IN (?)") {
		t.Fatalf("branch ids were not expanded: %s", query)
	}
	if !strings.Contains(query, "branch_id IN ($5,$6)") {
		t.Fatalf("expected expanded branch id placeholders, got: %s", query)
	}
	if !strings.Contains(query, "branch_id IS NULL") {
		t.Fatalf("expected null branch scope, got: %s", query)
	}
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %#v", len(args), args)
	}
}
