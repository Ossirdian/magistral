package prometheus

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	httpBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	grpcBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}
	store       = newMetricStore()
)

type metricStore struct {
	mu sync.RWMutex

	httpCount map[string]uint64
	httpHist  map[string]*histState

	grpcCount map[string]uint64
	grpcHist  map[string]*histState
}

type histState struct {
	Buckets []uint64
	Count   uint64
	Sum     float64
}

func newMetricStore() *metricStore {
	return &metricStore{
		httpCount: make(map[string]uint64),
		httpHist:  make(map[string]*histState),
		grpcCount: make(map[string]uint64),
		grpcHist:  make(map[string]*histState),
	}
}

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(renderMetrics()))
	})
}

func RecordHTTPRequest(method, path, status string, durationSec float64) {
	key := metricKey(method, path, status)

	store.mu.Lock()
	defer store.mu.Unlock()

	store.httpCount[key]++
	observe(store.httpHist, key, httpBuckets, durationSec)
}

func RecordGRPCRequest(method, code string, durationSec float64) {
	key := metricKey(method, code)

	store.mu.Lock()
	defer store.mu.Unlock()

	store.grpcCount[key]++
	observe(store.grpcHist, key, grpcBuckets, durationSec)
}

func observe(target map[string]*histState, key string, buckets []float64, value float64) {
	state, ok := target[key]
	if !ok {
		state = &histState{Buckets: make([]uint64, len(buckets))}
		target[key] = state
	}

	for i, bound := range buckets {
		if value <= bound {
			state.Buckets[i]++
		}
	}
	state.Count++
	state.Sum += value
}

func renderMetrics() string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var b strings.Builder

	b.WriteString("# HELP http_requests_total Total number of HTTP requests.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	for _, key := range sortedKeys(store.httpCount) {
		parts := strings.Split(key, "\x00")
		fmt.Fprintf(
			&b,
			"http_requests_total{method=%q,path=%q,status=%q} %d\n",
			escapeLabel(parts[0]),
			escapeLabel(parts[1]),
			escapeLabel(parts[2]),
			store.httpCount[key],
		)
	}

	renderHistogram(&b, "http_request_duration_seconds", "HTTP request duration.", store.httpHist, httpBuckets, []string{"method", "path", "status"})

	b.WriteString("# HELP grpc_requests_total Total number of gRPC requests.\n")
	b.WriteString("# TYPE grpc_requests_total counter\n")
	for _, key := range sortedKeys(store.grpcCount) {
		parts := strings.Split(key, "\x00")
		fmt.Fprintf(
			&b,
			"grpc_requests_total{method=%q,code=%q} %d\n",
			escapeLabel(parts[0]),
			escapeLabel(parts[1]),
			store.grpcCount[key],
		)
	}

	renderHistogram(&b, "grpc_request_duration_seconds", "gRPC request duration.", store.grpcHist, grpcBuckets, []string{"method", "code"})

	return b.String()
}

func renderHistogram(
	b *strings.Builder,
	name string,
	help string,
	values map[string]*histState,
	buckets []float64,
	labelOrder []string,
) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s histogram\n", name)

	for _, key := range sortedKeysHist(values) {
		parts := strings.Split(key, "\x00")
		state := values[key]

		for i, bound := range buckets {
			fmt.Fprintf(b, "%s_bucket{%s,le=%q} %d\n", name, labels(labelOrder, parts), strconv.FormatFloat(bound, 'f', -1, 64), state.Buckets[i])
		}
		fmt.Fprintf(b, "%s_bucket{%s,le=\"+Inf\"} %d\n", name, labels(labelOrder, parts), state.Count)
		fmt.Fprintf(b, "%s_sum{%s} %s\n", name, labels(labelOrder, parts), strconv.FormatFloat(state.Sum, 'f', -1, 64))
		fmt.Fprintf(b, "%s_count{%s} %d\n", name, labels(labelOrder, parts), state.Count)
	}
}

func labels(labelOrder, values []string) string {
	parts := make([]string, 0, len(labelOrder))
	for i, label := range labelOrder {
		if i >= len(values) {
			break
		}
		parts = append(parts, fmt.Sprintf(`%s=%q`, label, escapeLabel(values[i])))
	}
	return strings.Join(parts, ",")
}

func metricKey(parts ...string) string {
	return strings.Join(parts, "\x00")
}

func sortedKeys(m map[string]uint64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysHist(m map[string]*histState) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}
