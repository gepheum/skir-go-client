package skir_client

import (
	"encoding/json"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers: a minimal Color enum for case-compatibility tests
// ─────────────────────────────────────────────────────────────────────────────

type colorKind int

const (
	colorKindUnknown colorKind = 0
	colorKindRed     colorKind = 1
	colorKindGreen   colorKind = 2
	colorKindCustom  colorKind = 3
)

type testColor struct {
	kind         colorKind
	unrecognized *Internal__UnrecognizedVariant
	customValue  int32
}

var (
	testColorUnknown = testColor{kind: colorKindUnknown}
	testColorRed     = testColor{kind: colorKindRed}
	testColorGreen   = testColor{kind: colorKindGreen}
)

func newTestColorEnumAdapter() *Internal__EnumAdapter[testColor] {
	a := NewEnumAdapter[testColor](
		"test", "Color", "",
		func(c testColor) int { return int(c.kind) },
		4,
		testColorUnknown,
		func(u *Internal__UnrecognizedVariant) testColor {
			return testColor{kind: colorKindUnknown, unrecognized: u}
		},
		func(c testColor) *Internal__UnrecognizedVariant { return c.unrecognized },
	)
	// Register with lower_case names (as the future generator will produce).
	a.AddConstantVariant(1, "red", int(colorKindRed), "", testColorRed)
	a.AddConstantVariant(2, "green", int(colorKindGreen), "", testColorGreen)
	Internal__AddWrapperVariant(
		a, 3, "custom", int(colorKindCustom),
		Int32Serializer(),
		"",
		func(v int32) testColor { return testColor{kind: colorKindCustom, customValue: v} },
		func(c testColor) int32 { return c.customValue },
	)
	a.Finalize()
	return a
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

// TestEnumNameCase_SerializeLowerCase verifies that a constant registered with
// a lower_case name is serialized to lower_case in readable (non-nil eolIndent)
// JSON.
func TestEnumNameCase_SerializeLowerCase(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	got := ser.ToJson(testColorRed, Readable{})
	// Should be the JSON string "red" (with quotes).
	if got != `"red"` {
		t.Errorf("ToJson(RED, readable) = %q, want %q", got, `"red"`)
	}
}

// TestEnumNameCase_ParseUpperCase verifies that UPPER_CASE constant names
// (produced by old serializers) are correctly parsed.
func TestEnumNameCase_ParseUpperCase(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	// Parse "RED" (uppercase) – produced by old serializers.
	result, err := ser.FromJson(`"RED"`)
	if err != nil {
		t.Fatalf("FromJson(%q): %v", `"RED"`, err)
	}
	if result.kind != colorKindRed {
		t.Errorf("FromJson(%q).kind = %v, want %v", `"RED"`, result.kind, colorKindRed)
	}
}

// TestEnumNameCase_ParseLowerCase verifies that lower_case constant names
// (produced by new serializers) are correctly parsed.
func TestEnumNameCase_ParseLowerCase(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	result, err := ser.FromJson(`"green"`)
	if err != nil {
		t.Fatalf("FromJson(%q): %v", `"green"`, err)
	}
	if result.kind != colorKindGreen {
		t.Errorf("FromJson(%q).kind = %v, want %v", `"green"`, result.kind, colorKindGreen)
	}
}

// TestEnumNameCase_UpperAndLowerYieldSameResult verifies that parsing the same
// constant name in either casing yields identical results.
func TestEnumNameCase_UpperAndLowerYieldSameResult(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	fromUpper, err := ser.FromJson(`"RED"`)
	if err != nil {
		t.Fatalf("FromJson(RED): %v", err)
	}
	fromLower, err := ser.FromJson(`"red"`)
	if err != nil {
		t.Fatalf("FromJson(red): %v", err)
	}
	if fromUpper.kind != fromLower.kind {
		t.Errorf("Parsing RED and red yield different kinds: %v vs %v", fromUpper.kind, fromLower.kind)
	}
	if fromUpper.kind != colorKindRed {
		t.Errorf("Expected colorKindRed, got %v", fromUpper.kind)
	}
}

// TestEnumNameCase_WrapperUpperCaseKind verifies that a readable JSON object
// with UPPER_CASE "kind" field is parsed correctly.
func TestEnumNameCase_WrapperUpperCaseKind(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	// Old serializer output: {"kind":"CUSTOM","value":42}
	upperJSON, _ := json.Marshal(map[string]interface{}{"kind": "CUSTOM", "value": 42})
	lowerJSON, _ := json.Marshal(map[string]interface{}{"kind": "custom", "value": 42})

	fromUpper, err := ser.FromJson(string(upperJSON))
	if err != nil {
		t.Fatalf("FromJson(UPPER): %v", err)
	}
	fromLower, err := ser.FromJson(string(lowerJSON))
	if err != nil {
		t.Fatalf("FromJson(lower): %v", err)
	}
	if fromUpper.kind != colorKindCustom {
		t.Errorf("FromJson(UPPER).kind = %v, want custom", fromUpper.kind)
	}
	if fromLower.kind != colorKindCustom {
		t.Errorf("FromJson(lower).kind = %v, want custom", fromLower.kind)
	}
	if fromUpper.customValue != fromLower.customValue {
		t.Errorf("custom values differ: %v vs %v", fromUpper.customValue, fromLower.customValue)
	}
}

// TestEnumNameCase_WrapperSerializesLowerCase verifies that the wrapper variant
// serializes to lower_case "kind" in readable JSON.
func TestEnumNameCase_WrapperSerializesLowerCase(t *testing.T) {
	a := newTestColorEnumAdapter()
	ser := a.Serializer()

	custom := testColor{kind: colorKindCustom, customValue: 7}
	got := ser.ToJson(custom, Readable{})
	// Should contain "kind": "custom" (lowercase).
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(got), &obj); err != nil {
		t.Fatalf("unmarshal readable JSON: %v", err)
	}
	if obj["kind"] != "custom" {
		t.Errorf("readable JSON kind = %q, want %q", obj["kind"], "custom")
	}
}
