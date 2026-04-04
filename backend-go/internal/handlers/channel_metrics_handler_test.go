package handlers

import (
	"testing"
	"time"

	"github.com/BenedictKing/ccx/internal/metrics"
)

type fakePersistenceStore struct {
	bucketsByMetricsKey map[string][]metrics.AggregatedBucket
}

func (f *fakePersistenceStore) AddRecord(record metrics.PersistentRecord) {}
func (f *fakePersistenceStore) LoadRecords(since time.Time, apiType string) ([]metrics.PersistentRecord, error) {
	return nil, nil
}
func (f *fakePersistenceStore) LoadLatestTimestamps(apiType string) (map[string]*metrics.KeyLatestTimestamps, error) {
	return nil, nil
}
func (f *fakePersistenceStore) QueryAggregatedHistory(apiType string, since time.Time, intervalSeconds int64, metricsKey string, baseURL string) ([]metrics.AggregatedBucket, error) {
	return append([]metrics.AggregatedBucket(nil), f.bucketsByMetricsKey[metricsKey]...), nil
}
func (f *fakePersistenceStore) CleanupOldRecords(before time.Time) (int64, error) { return 0, nil }
func (f *fakePersistenceStore) DeleteRecordsByMetricsKeys(metricsKeys []string, apiType string) (int64, error) {
	return 0, nil
}
func (f *fakePersistenceStore) Close() error { return nil }

func TestFilterBucketsByURLsIsolatesChannelsByMetricsKey(t *testing.T) {
	baseURL := "https://shared.example.com"
	keyA := "sk-a"
	keyB := "sk-b"

	store := &fakePersistenceStore{
		bucketsByMetricsKey: map[string][]metrics.AggregatedBucket{
			metrics.GenerateMetricsKey(baseURL, keyA): {
				{Timestamp: time.Unix(3600, 0), TotalRequests: 1, SuccessCount: 1},
			},
			metrics.GenerateMetricsKey(baseURL, keyB): {
				{Timestamp: time.Unix(3600, 0), TotalRequests: 2, SuccessCount: 1},
			},
		},
	}

	channelABuckets := filterBucketsByURLs(store, "messages", time.Unix(0, 0), 3600, []string{baseURL}, []string{keyA})
	channelBBuckets := filterBucketsByURLs(store, "messages", time.Unix(0, 0), 3600, []string{baseURL}, []string{keyB})

	if len(channelABuckets) != 1 || channelABuckets[0].TotalRequests != 1 {
		t.Fatalf("channel A buckets = %+v, want only keyA data", channelABuckets)
	}
	if len(channelBBuckets) != 1 || channelBBuckets[0].TotalRequests != 2 {
		t.Fatalf("channel B buckets = %+v, want only keyB data", channelBBuckets)
	}
}
