package parser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/weaweawe01/velocity-ast/internal/dump"
)

func TestExploitChainMatchesJavaBaseline(t *testing.T) {
	cases := []string{
		"exploit_chain",
		"foreach_basic",
		"macro_define_call",
		"if_basic",
		"if_else",
		"if_elseif_else",
		"if_eq",
		"if_ne",
		"if_gt",
		"if_lt",
		"if_ge",
		"if_le",
		"if_and",
		"if_or",
		"if_keyword_ops",
		"if_not",
		"if_not_paren_eq",
		"if_true_false",
		"set_add",
		"set_array",
		"set_sub",
		"set_mul",
		"set_div",
		"set_float",
		"set_neg_float",
		"set_neg_ref",
		"set_range",
		"set_map",
		"set_index_int",
		"set_index_ref",
		"set_prop_ref",
		"set_formal_ref",
		"set_formal_ref_alt",
		"set_formal_prop_ref",
		"set_paren_mul",
		"braced_if",
		"set_space",
		"set_brace",
		"foreach_else",
		"quiet_ref",
		"comment_line",
		"comment_block",
		"textblock",
		"macro_default",
		"escape_double",
		"define_basic",
		"parse_basic",
		"include_multi",
		"evaluate_ref",
		"stop_directive",
		"break_directive",
		"macro_call_no_paren",
		"parse_newline",
		"stop_newline",
		"break_newline",
		"parse_then_text",
		"include_then_text",
		"stop_then_text",
		"break_then_text",
		"if_newline_body",
	}

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			assertCaseMatchesJavaBaseline(t, name)
		})
	}
}

func assertCaseMatchesJavaBaseline(t *testing.T, caseName string) {
	rootDir := moduleRoot(t)
	casePath := filepath.Join(rootDir, "testdata", "cases", caseName+".vtl")
	expectedPath := filepath.Join(rootDir, "testdata", "expected-java", caseName+".ast")

	tplBytes, err := os.ReadFile(casePath)
	if err != nil {
		t.Fatalf("read case: %v", err)
	}
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}

	root, tokens, err := Parse(string(tplBytes))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	got := dump.Render(root, tokens)
	expected := string(expectedBytes)
	if got != expected {
		t.Fatalf("java/go ast mismatch\n--- got ---\n%s\n--- expected ---\n%s", got, expected)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/parser/compat_test.go -> module root
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
