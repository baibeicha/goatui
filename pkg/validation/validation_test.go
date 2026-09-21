package validation

import (
	"strings"
	"testing"
)

func TestValidation_Rules(t *testing.T) {
	// 1. Required
	req := Required()
	if ok, _ := req(""); ok {
		t.Error("expected Required to fail on empty string")
	}
	if ok, _ := req("   "); ok {
		t.Error("expected Required to fail on whitespace string")
	}
	if ok, _ := req("hello"); !ok {
		t.Error("expected Required to pass on valid string")
	}

	// Custom message
	customReq := Required("Custom required message")
	if ok, msg := customReq(""); ok || msg != "Custom required message" {
		t.Errorf("expected custom message, got %s", msg)
	}

	// 2. MinLength & MaxLength
	min3 := MinLength(3)
	if ok, _ := min3("ab"); ok {
		t.Error("expected MinLength(3) to fail for 'ab'")
	}
	if ok, _ := min3("abc"); !ok {
		t.Error("expected MinLength(3) to pass for 'abc'")
	}
	// Unicode rune test
	if ok, _ := min3("при"); !ok {
		t.Error("expected MinLength(3) to count unicode runes properly")
	}

	max5 := MaxLength(5)
	if ok, _ := max5("12345"); !ok {
		t.Error("expected MaxLength(5) to pass for '12345'")
	}
	if ok, _ := max5("123456"); ok {
		t.Error("expected MaxLength(5) to fail for '123456'")
	}

	// LengthRange
	lr := LengthRange(2, 4)
	if ok, _ := lr("a"); ok {
		t.Error("expected LengthRange to fail on too short")
	}
	if ok, _ := lr("123"); !ok {
		t.Error("expected LengthRange to pass on 3 chars")
	}
	if ok, _ := lr("12345"); ok {
		t.Error("expected LengthRange to fail on too long")
	}

	// 3. IntRange & FloatRange
	ir := IntRange(10, 100)
	if ok, _ := ir("5"); ok {
		t.Error("expected IntRange to fail for 5")
	}
	if ok, _ := ir("10"); !ok {
		t.Error("expected IntRange to pass for 10")
	}
	if ok, _ := ir("100"); !ok {
		t.Error("expected IntRange to pass for 100")
	}
	if ok, _ := ir("101"); ok {
		t.Error("expected IntRange to fail for 101")
	}
	if ok, _ := ir("abc"); ok {
		t.Error("expected IntRange to fail for non-int")
	}

	fr := FloatRange(0.5, 9.5)
	if ok, _ := fr("0.2"); ok {
		t.Error("expected FloatRange to fail for 0.2")
	}
	if ok, _ := fr("3.1415"); !ok {
		t.Error("expected FloatRange to pass for 3.1415")
	}
	if ok, _ := fr("10.0"); ok {
		t.Error("expected FloatRange to fail for 10.0")
	}

	// 4. Regex
	reg := Regex(`^[A-Z]{3}-\d{3}$`)
	if ok, _ := reg("ABC-123"); !ok {
		t.Error("expected Regex to pass for ABC-123")
	}
	if ok, _ := reg("abc-123"); ok {
		t.Error("expected Regex to fail for lowercase")
	}

	// 5. Email
	em := Email()
	if ok, _ := em("user@example.com"); !ok {
		t.Error("expected Email to pass for user@example.com")
	}
	if ok, _ := em("invalid-email"); ok {
		t.Error("expected Email to fail for invalid-email")
	}
	if ok, _ := em("@missing-user.com"); ok {
		t.Error("expected Email to fail for @missing-user.com")
	}

	// 6. URL
	u := URL()
	if ok, _ := u("https://golang.org"); !ok {
		t.Error("expected URL to pass for https://golang.org")
	}
	if ok, _ := u("http://localhost:8080/api"); !ok {
		t.Error("expected URL to pass for http://localhost:8080/api")
	}
	if ok, _ := u("not_a_url"); ok {
		t.Error("expected URL to fail for not_a_url")
	}

	// 7. Numeric, Alpha, Alphanumeric
	num := Numeric()
	if ok, _ := num("123456"); !ok {
		t.Error("expected Numeric to pass for digits")
	}
	if ok, _ := num("123a"); ok {
		t.Error("expected Numeric to fail for letters")
	}

	alpha := Alpha()
	if ok, _ := alpha("HelloWorld"); !ok {
		t.Error("expected Alpha to pass for letters")
	}
	if ok, _ := alpha("Hello 123"); ok {
		t.Error("expected Alpha to fail for digits and spaces")
	}

	alphanum := Alphanumeric()
	if ok, _ := alphanum("Go123"); !ok {
		t.Error("expected Alphanumeric to pass for letters and digits")
	}
	if ok, _ := alphanum("Go_123!"); ok {
		t.Error("expected Alphanumeric to fail for punctuation")
	}

	// 8. Custom
	cust := Custom(func(val string) (bool, string) {
		if strings.HasPrefix(val, "GOAT_") {
			return true, ""
		}
		return false, "Must start with GOAT_"
	})
	if ok, _ := cust("GOAT_FAST"); !ok {
		t.Error("expected Custom to pass for GOAT_FAST")
	}
	if ok, msg := cust("SLOW"); ok || msg != "Must start with GOAT_" {
		t.Errorf("expected Custom to fail with specific message, got %s", msg)
	}
}

func TestValidation_ValidatorAndForm(t *testing.T) {
	// Validator chaining
	v := New(Required(), MinLength(4), Alphanumeric())

	// Empty
	res := v.Validate("")
	if res.IsValid() {
		t.Error("expected empty string to be invalid")
	}
	if res.Error() == "" {
		t.Error("expected error message")
	}

	// Valid
	res = v.Validate("Admin1")
	if !res.IsValid() {
		t.Fatalf("expected 'Admin1' to be valid, got errors: %v", res.AllErrors())
	}

	// Form validation
	form := NewForm().
		Field("username", Required(), MinLength(3), Alphanumeric()).
		Field("email", Required(), Email()).
		Field("age", Required(), IntRange(18, 120))

	// Test invalid form
	formData := map[string]string{
		"username": "ab",
		"email":    "bad-email",
		"age":      "15",
	}
	formRes := form.Validate(formData)
	if formRes.IsValid() {
		t.Fatal("expected form to be invalid")
	}
	if formRes.Error("username") == "" {
		t.Error("expected username error")
	}
	if formRes.Error("email") == "" {
		t.Error("expected email error")
	}
	if formRes.Error("age") == "" {
		t.Error("expected age error")
	}

	// Test valid form
	formDataValid := map[string]string{
		"username": "superadmin",
		"email":    "admin@goatui.dev",
		"age":      "25",
	}
	formResValid := form.Validate(formDataValid)
	if !formResValid.IsValid() {
		t.Fatalf("expected form to be valid, got errors: %v", formResValid.Fields)
	}
}

func TestValidation_Optional(t *testing.T) {
	rule := Optional(Email())

	// Empty and whitespace should pass
	if ok, _ := rule(""); !ok {
		t.Error("expected Optional(Email) to pass for empty string")
	}
	if ok, _ := rule("   "); !ok {
		t.Error("expected Optional(Email) to pass for whitespace")
	}

	// Non-empty valid email passes
	if ok, _ := rule("test@example.com"); !ok {
		t.Error("expected Optional(Email) to pass for valid email")
	}

	// Non-empty invalid email fails
	if ok, msg := rule("invalid"); ok || msg == "" {
		t.Error("expected Optional(Email) to fail for invalid email")
	}
}
