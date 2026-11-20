package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func goos() string {
	return runtime.GOOS
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func nonEmpty(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		out[v] = true
	}
	return out
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, strings.TrimSpace(v))
	}
	return nonNilStrings(out)
}

func emptyStrings() []string {
	return []string{}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func contains(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}

func removeString(values []string, needle string) []string {
	var out []string
	for _, v := range values {
		if v != needle {
			out = append(out, v)
		}
	}
	return out
}

func mergeCounts(left, right map[string]int) {
	for k, v := range right {
		left[k] += v
	}
}

func floatValue(v any) float64 {
	f, _ := numberValue(v)
	return f
}

func numberValue(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func intValue(v any, fallback int) int {
	if f, ok := numberValue(v); ok {
		return int(f)
	}
	return fallback
}

func truthy(v any, fallback bool) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		if x == "" {
			return fallback
		}
		return x == "1" || strings.EqualFold(x, "true") || strings.EqualFold(x, "yes")
	default:
		if v == nil {
			return fallback
		}
		return fallback
	}
}

func anySlice(v any) []any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []any:
		return x
	case []map[string]any:
		out := make([]any, len(x))
		for i, v := range x {
			out[i] = v
		}
		return out
	case []string:
		out := make([]any, len(x))
		for i, v := range x {
			out[i] = v
		}
		return out
	default:
		return nil
	}
}

func mapSlice(v any) []map[string]any {
	var out []map[string]any
	for _, item := range anySlice(v) {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func signalSlice(v any) []map[string]any { return mapSlice(v) }

func toStringSlice(v any) []string {
	if s, ok := v.(string); ok {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return []string{s}
	}
	var out []string
	for _, item := range anySlice(v) {
		if s := stringValue(item); strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func valueOrEmptyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok && m != nil {
		return m
	}
	return map[string]any{}
}

func valueOrEmptyArray(v any) []any {
	if arr := anySlice(v); arr != nil {
		return arr
	}
	return []any{}
}

func firstArray(m map[string]any, keys ...string) []any {
	for _, key := range keys {
		if arr := anySlice(m[key]); len(arr) > 0 {
			return arr
		}
	}
	return nil
}

func firstAny(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok && v != nil && stringValue(v) != "" {
			return v
		}
	}
	return nil
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func firstN[T any](values []T, n int) []T {
	if len(values) > n {
		return values[:n]
	}
	return values
}

func mapKeys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func cloneMap(m map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sortedCountMap(m map[string]int) map[string]int {
	return m
}

func joinCounts(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func joinAnyCounts(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%v", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

func splitFields(raw string) []string {
	var out []string
	for _, part := range regexp.MustCompile(`[\n,]`).Split(raw, -1) {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return nonNilStrings(out)
}

func reorderIntermixedFlags(args []string, valueFlags map[string]bool) []string {
	var flagsOut []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positional = append(positional, arg)
			continue
		}
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		flagsOut = append(flagsOut, arg)
		if strings.Contains(arg, "=") {
			continue
		}
		if valueFlags[name] && i+1 < len(args) {
			i++
			flagsOut = append(flagsOut, args[i])
		}
	}
	return append(flagsOut, positional...)
}

func boolSlice(ok bool, value string) []string {
	if ok {
		return []string{value}
	}
	return nil
}

func roundN(v float64, n int) float64 {
	p := math.Pow10(n)
	return math.Round(v*p) / p
}

func round1(v float64) float64 { return roundN(v, 1) }

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sumInts(values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

func lastInts(values []int, n int) []int {
	if len(values) <= n {
		return values
	}
	return values[len(values)-n:]
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return v
	}
	return fallback
}

// datePrefix returns the leading YYYY-MM-DD portion of a date/datetime string
// without panicking on short or empty input (time.Parse then reports the error).
func datePrefix(date string) string {
	date = strings.TrimSpace(date)
	if len(date) > 10 {
		return date[:10]
	}
	return date
}

func dateToUnix(date string) int64 {
	t, err := time.Parse("2006-01-02", datePrefix(date))
	if err != nil {
		return 0
	}
	return t.Unix()
}

func lookbackHours(fromDate, toDate string) int {
	start, err1 := time.Parse("2006-01-02", datePrefix(fromDate))
	end, err2 := time.Parse("2006-01-02", datePrefix(toDate))
	if err1 != nil || err2 != nil {
		return 168
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days*24 < 24 {
		return 24
	}
	return days * 24
}

func tbsForRange(fromDate, toDate string) string {
	start, err1 := time.Parse("2006-01-02", datePrefix(fromDate))
	end, err2 := time.Parse("2006-01-02", datePrefix(toDate))
	if err1 != nil || err2 != nil {
		return "qdr:w"
	}
	days := int(end.Sub(start).Hours() / 24)
	if days <= 1 {
		return "qdr:d"
	}
	if days <= 7 {
		return "qdr:w"
	}
	return "qdr:m"
}

func normalizeLooseDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) >= 10 && value[4] == '-' && value[7] == '-' {
		return value
	}
	rel := regexp.MustCompile(`(?i)^\s*(\d+)\s+(minute|minutes|hour|hours|day|days|week|weeks|month|months)\s+ago\s*$`).FindStringSubmatch(value)
	if len(rel) > 0 {
		n, _ := strconv.Atoi(rel[1])
		unit := strings.ToLower(rel[2])
		d := time.Duration(0)
		switch {
		case strings.HasPrefix(unit, "minute"):
			d = time.Duration(n) * time.Minute
		case strings.HasPrefix(unit, "hour"):
			d = time.Duration(n) * time.Hour
		case strings.HasPrefix(unit, "day"):
			d = time.Duration(n) * 24 * time.Hour
		case strings.HasPrefix(unit, "week"):
			d = time.Duration(n) * 7 * 24 * time.Hour
		default:
			d = time.Duration(n) * 30 * 24 * time.Hour
		}
		return time.Now().UTC().Add(-d).Format(time.RFC3339Nano)
	}
	if parsed, err := http.ParseTime(value); err == nil {
		return parsed.UTC().Format(time.RFC3339Nano)
	}
	return value
}

func isoZ(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func cleanError(text string) string {
	return regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`).ReplaceAllString(strings.TrimSpace(text), "")
}

func mustRegexes(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		out[i] = regexp.MustCompile(pattern)
	}
	return out
}

type cancelFunc func()

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func contextDeadlineExceeded() error { return context.DeadlineExceeded }
