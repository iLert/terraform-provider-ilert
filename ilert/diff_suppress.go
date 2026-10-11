package ilert

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The API does not keep a text template the way it was sent. It parses the template into
// elements, stores only those and rebuilds the text from them on every read, so a template
// can come back spaced differently than it was written: `{{ subject.splitTakeAt(" ", 0) }}`
// reads back as `{{ subject.splitTakeAt(" ",0) }}`. Both are stored as the same elements,
// but the plain string comparison saw a change on every plan that no apply could settle
// (#165). suppressEquivalentTemplateDiff hides the spacing part of that difference.
//
// It ports the two places where the server's parser drops whitespace (TemplateDSLEngine in
// the monolith, verified 09.10.2026): the trim of what is between the braces, and the
// tokenizer skipping spaces in the arguments of a function call. It reduces a template to
// the parts that parser actually reads, so two templates with the same key are stored as
// the same elements. The rest of the server's rewriting, such as the quotes it adds around
// bare arguments or the trim it applies to numeric arguments, is not modelled. Those
// differences keep showing up as a diff: the port may miss an equivalence, but it never
// treats a real change as one.
func suppressEquivalentTemplateDiff(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldKey, ok := templateKey(oldValue)
	if !ok {
		return false
	}
	newKey, ok := templateKey(newValue)
	return ok && slices.Equal(oldKey, newKey)
}

var (
	// The server's PADS_REGEX `\{\{.*?}}`. Java's `.` stops at every line terminator while
	// Go's only stops at `\n`, so the character class spells out Java's set.
	templateVariablePattern = regexp.MustCompile(`\{\{[^\n\r\x{85}\x{2028}\x{2029}]*?\}\}`)

	// The server's NEW_FUNCTION_PATTERN `^(?!.*##)[a-zA-Z0-9_.#^\[\]]+\(\s*.*\s*\)$` with
	// Java's `\s` and `.` spelled out. RE2 has no lookahead, so the `##` check is done
	// separately in templateVariableKey.
	templateFunctionPattern = regexp.MustCompile(`^[a-zA-Z0-9_.#^\[\]]+\([ \t\n\x0B\f\r]*[^\n\r\x{85}\x{2028}\x{2029}]*[ \t\n\x0B\f\r]*\)$`)
)

// templateKey reduces a template to what the server's parser reads from it. It reports
// false for a template the server would reject or that the port does not cover, in which
// case the diff is left alone.
func templateKey(text string) ([]string, bool) {
	if javaTrim(text) == "" {
		return nil, false
	}

	var key []string
	start := 0
	for _, match := range templateVariablePattern.FindAllStringIndex(text, -1) {
		if match[0] > start {
			key = append(key, "text:"+text[start:match[0]])
		}
		variable, ok := templateVariableKey(text[match[0]+2 : match[1]-2])
		if !ok {
			return nil, false
		}
		key = append(key, variable)
		start = match[1]
	}
	if start < len(text) {
		key = append(key, "text:"+text[start:])
	}

	return key, true
}

// templateVariableKey mirrors how the server reads what is between `{{` and `}}`. A
// function call such as `subject.splitTakeAt(" ", 0)` goes through a tokenizer that skips
// plain spaces. Anything else, a plain variable or the legacy `##` syntax, is read by
// a parser that keeps them, so `{{ subject . splitTakeAt(" ",0) }}` is a single variable
// named `subject . splitTakeAt(" ",0)` and must not compare equal to the function call.
func templateVariableKey(raw string) (string, bool) {
	inner := javaTrim(raw)
	if inner == "" {
		return "", false // the server rejects an empty variable
	}

	if strings.Contains(inner, "##") || !templateFunctionPattern.MatchString(inner) {
		return "verbatim:" + inner, true
	}

	stripped, ok := stripIgnoredSpaces(inner)
	if !ok {
		return "", false
	}
	return "function:" + stripped, true
}

// stripIgnoredSpaces drops the spaces the server's tokenizer (TemplateDSLEngine.tokenize)
// skips: every plain space that is neither escaped nor inside a quoted argument. Skipping
// one changes none of the tokenizer's state, so the tokens it builds depend only on what
// is left. The state machine follows the tokenizer branch for branch.
func stripIgnoredSpaces(expression string) (string, bool) {
	var b strings.Builder
	escaped, embraced, quoted := false, false, false
	for _, c := range expression {
		switch {
		case c == '\\' && !escaped:
			escaped = true
		case escaped:
			escaped = false
		case quoted:
			if c == '"' {
				quoted = false
			}
		case c == '"' && embraced:
			quoted = true
		case c == '(':
			if embraced {
				return "", false // nested parentheses, rejected by the server
			}
			embraced = true
		case c == ')' && embraced:
			embraced = false
		case c == ' ':
			continue
		}
		b.WriteRune(c)
	}

	if quoted || embraced {
		return "", false // an unclosed string or parenthesis, rejected by the server
	}
	return b.String(), true
}

// The API lowercases a service alias and trims it (Service.setAlias in the monolith,
// verified 09.10.2026), so an alias written with capitals or surrounding spaces read back
// different from the configuration on every plan (#165). The configured value is
// normalized the same way and compared against the one the API returned. Only ASCII is
// normalized: Go and Java lowercase some other characters differently (the Greek final
// sigma, for one), and suppressing on a mismatch could hide a real change.
func suppressEquivalentServiceAliasDiff(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	for i := 0; i < len(newValue); i++ {
		if newValue[i] >= utf8.RuneSelf {
			return false
		}
	}
	return oldValue == strings.ToLower(javaTrim(newValue))
}

// javaTrim matches Java's String.trim, which strips every character up to and including
// the space from both ends, control characters too, rather than Unicode whitespace.
func javaTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}
