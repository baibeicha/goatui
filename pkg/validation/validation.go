package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Result represents the outcome of a validation execution.
type Result struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// OK returns a successful validation result.
func OK() Result {
	return Result{Valid: true, Errors: nil}
}

// Fail returns a failed validation result with one or more error messages.
func Fail(msgs ...string) Result {
	var errs []string
	for _, m := range msgs {
		if m != "" {
			errs = append(errs, m)
		}
	}
	if len(errs) == 0 {
		errs = []string{"Validation failed"}
	}
	return Result{
		Valid:  false,
		Errors: errs,
	}
}

// IsValid returns true if all validation checks passed.
func (r Result) IsValid() bool {
	return r.Valid
}

// Error returns the first error message, or an empty string if valid.
func (r Result) Error() string {
	if len(r.Errors) > 0 {
		return r.Errors[0]
	}
	return ""
}

// FirstError returns the first error message, or an empty string if valid.
func (r Result) FirstError() string {
	return r.Error()
}

// AllErrors returns all error messages generated during validation.
func (r Result) AllErrors() []string {
	return r.Errors
}

// Rule defines a validation predicate for a string value.
// Returns (true, "") if valid, or (false, errorMsg) if invalid.
type Rule func(val string) (bool, string)

// Required ensures that the value is not empty and not just whitespace.
func Required(customMsg ...string) Rule {
	msg := "This field is required"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		if strings.TrimSpace(val) == "" {
			return false, msg
		}
		return true, ""
	}
}

// MinLength ensures the value contains at least min unicode characters.
func MinLength(min int, customMsg ...string) Rule {
	msg := fmt.Sprintf("Must be at least %d characters", min)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		if utf8.RuneCountInString(val) < min {
			return false, msg
		}
		return true, ""
	}
}

// MaxLength ensures the value contains at most max unicode characters.
func MaxLength(max int, customMsg ...string) Rule {
	msg := fmt.Sprintf("Must be at most %d characters", max)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		if utf8.RuneCountInString(val) > max {
			return false, msg
		}
		return true, ""
	}
}

// LengthRange ensures the character length is between min and max inclusive.
func LengthRange(min, max int, customMsg ...string) Rule {
	msg := fmt.Sprintf("Length must be between %d and %d characters", min, max)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		cnt := utf8.RuneCountInString(val)
		if cnt < min || cnt > max {
			return false, msg
		}
		return true, ""
	}
}

// IntRange ensures the value parses to an integer between min and max inclusive.
func IntRange(min, max int64, customMsg ...string) Rule {
	msg := fmt.Sprintf("Must be an integer between %d and %d", min, max)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n < min || n > max {
			return false, msg
		}
		return true, ""
	}
}

// FloatRange ensures the value parses to a floating point number between min and max inclusive.
func FloatRange(min, max float64, customMsg ...string) Rule {
	msg := fmt.Sprintf("Must be a number between %g and %g", min, max)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil || n < min || n > max {
			return false, msg
		}
		return true, ""
	}
}

// Regex ensures the value matches the given regular expression pattern.
func Regex(pattern string, customMsg ...string) Rule {
	re := regexp.MustCompile(pattern)
	msg := fmt.Sprintf("Must match pattern: %s", pattern)
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		if !re.MatchString(val) {
			return false, msg
		}
		return true, ""
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)

// Email ensures the value is a valid email address.
func Email(customMsg ...string) Rule {
	msg := "Must be a valid email address"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" || !emailRegex.MatchString(s) {
			return false, msg
		}
		return true, ""
	}
}

// URL ensures the value is a valid HTTP/HTTPS URL.
func URL(customMsg ...string) Rule {
	msg := "Must be a valid URL"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		u, err := url.ParseRequestURI(s)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return false, msg
		}
		if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ftp" {
			return false, msg
		}
		return true, ""
	}
}

// Numeric ensures all characters in the string are digits.
func Numeric(customMsg ...string) Rule {
	msg := "Must contain only digits"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		for _, r := range s {
			if !unicode.IsDigit(r) {
				return false, msg
			}
		}
		return true, ""
	}
}

// Alpha ensures all characters are letters.
func Alpha(customMsg ...string) Rule {
	msg := "Must contain only letters"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false, msg
			}
		}
		return true, ""
	}
}

// Alphanumeric ensures all characters are letters or digits.
func Alphanumeric(customMsg ...string) Rule {
	msg := "Must contain only letters and numbers"
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	return func(val string) (bool, string) {
		s := strings.TrimSpace(val)
		if s == "" {
			return false, msg
		}
		for _, r := range s {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return false, msg
			}
		}
		return true, ""
	}
}

// Custom accepts any custom validation logic.
func Custom(fn func(val string) (bool, string)) Rule {
	return fn
}

// Optional wraps a rule so that empty or whitespace-only values are considered valid.
// Non-empty values are delegated to the underlying rule.
func Optional(rule Rule) Rule {
	return func(val string) (bool, string) {
		if strings.TrimSpace(val) == "" {
			return true, ""
		}
		return rule(val)
	}
}

// Validator encapsulates a chain of rules for validating a single input field.
type Validator struct {
	rules []Rule
}

// New creates a new field Validator with optional initial rules.
func New(rules ...Rule) *Validator {
	return &Validator{rules: rules}
}

// Add appends validation rules to the field validator.
func (v *Validator) Add(rules ...Rule) *Validator {
	v.rules = append(v.rules, rules...)
	return v
}

// Validate evaluates all rules in sequence and collects any errors.
func (v *Validator) Validate(val string) Result {
	var errs []string
	for _, r := range v.rules {
		if ok, msg := r(val); !ok {
			errs = append(errs, msg)
		}
	}
	if len(errs) > 0 {
		return Result{Valid: false, Errors: errs}
	}
	return OK()
}

// Form manages declarative validation across multiple named form fields.
type Form struct {
	fields map[string]*Validator
}

// NewForm creates a new multi-field Form validator.
func NewForm() *Form {
	return &Form{
		fields: make(map[string]*Validator),
	}
}

// Field registers validation rules for a specific field name.
func (f *Form) Field(name string, rules ...Rule) *Form {
	if _, ok := f.fields[name]; !ok {
		f.fields[name] = New()
	}
	f.fields[name].Add(rules...)
	return f
}

// FormResult holds the validation outcome for an entire form.
type FormResult struct {
	Valid  bool              `json:"valid"`
	Fields map[string]Result `json:"fields"`
}

// IsValid returns true if all form fields passed validation.
func (fr FormResult) IsValid() bool {
	return fr.Valid
}

// FieldResult returns the Result for the specified field name.
func (fr FormResult) FieldResult(name string) Result {
	if r, ok := fr.Fields[name]; ok {
		return r
	}
	return OK()
}

// Error returns the first error message for a given field name, or empty string.
func (fr FormResult) Error(name string) string {
	return fr.FieldResult(name).Error()
}

// Validate executes validation on all registered form fields against the provided values.
func (f *Form) Validate(values map[string]string) FormResult {
	res := FormResult{
		Valid:  true,
		Fields: make(map[string]Result),
	}
	for name, v := range f.fields {
		val := values[name]
		r := v.Validate(val)
		res.Fields[name] = r
		if !r.IsValid() {
			res.Valid = false
		}
	}
	return res
}
