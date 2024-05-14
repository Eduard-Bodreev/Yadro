package xkcd

import (
	"runtime"
	"testing"
)

func BenchmarkFetchComicParallel(b *testing.B) {
	client := New()
	baseURL := "https://xkcd.com"
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.FetchComic(614, baseURL)
			if err != nil {
				b.Error("Error fetching comic: ", err)
			}
		}
	})
	b.Log("Number of goroutines after test: ", runtime.NumGoroutine())
}
