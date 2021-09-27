package cache

import (
	"github.com/rueian/rueidis/internal/proto"
	"strconv"
	"testing"
	"time"
)

const TTL = 100 * time.Millisecond
const Entries = 3

func TestLRU(t *testing.T) {

	setup := func(t *testing.T) *LRU {
		lru := NewLRU(EntryMinSize * Entries)
		if v := lru.GetOrPrepare("0", TTL); v.Type != 0 {
			t.Fatalf("got unexpected value from the first GetOrPrepare: %v", v)
		}
		lru.Update("0", proto.Message{Type: '+', String: "0"})
		return lru
	}

	t.Run("Cache Hit & Expire", func(t *testing.T) {
		lru := setup(t)
		if v := lru.GetOrPrepare("0", TTL); v.Type == 0 {
			t.Fatalf("did not get the value from the second GetOrPrepare")
		} else if v.String != "0" {
			t.Fatalf("got unexpected value from the second GetOrPrepare: %v", v)
		}
		time.Sleep(TTL)
		if v := lru.GetOrPrepare("0", TTL); v.Type != 0 {
			t.Fatalf("got unexpected value from the GetOrPrepare after ttl: %v", v)
		}
	})

	t.Run("Cache Miss", func(t *testing.T) {
		lru := setup(t)
		if v := lru.GetOrPrepare("1", TTL); v.Type != 0 {
			t.Fatalf("got unexpected value from the first GetOrPrepare: %v", v)
		}
	})

	t.Run("Cache Evict", func(t *testing.T) {
		lru := setup(t)
		for i := 1; i <= Entries; i++ {
			lru.GetOrPrepare(strconv.Itoa(i), TTL)
			lru.Update(strconv.Itoa(i), proto.Message{Type: '+', String: strconv.Itoa(i)})
		}
		if v := lru.GetOrPrepare("1", TTL); v.Type != 0 {
			t.Fatalf("got evicted value from the first GetOrPrepare: %v", v)
		}
		if v := lru.GetOrPrepare(strconv.Itoa(Entries), TTL); v.Type == 0 {
			t.Fatalf("did not get the latest value from the GetOrPrepare")
		} else if v.String != strconv.Itoa(Entries) {
			t.Fatalf("got unexpected value from the GetOrPrepare: %v", v)
		}
	})

	t.Run("Cache Delete", func(t *testing.T) {
		lru := setup(t)
		lru.Delete([]proto.Message{{String: "0"}})
		if v := lru.GetOrPrepare("0", TTL); v.Type != 0 {
			t.Fatalf("got unexpected value from the first GetOrPrepare: %v", v)
		}
	})

	t.Run("Cache DeleteAll", func(t *testing.T) {
		lru := setup(t)
		lru.DeleteAll()
		if v := lru.GetOrPrepare("0", TTL); v.Type != 0 {
			t.Fatalf("got unexpected value from the first GetOrPrepare: %v", v)
		}
	})
}

func BenchmarkLRU(b *testing.B) {
	lru := NewLRU(EntryMinSize * Entries)
	b.Run("GetOrPrepare", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				lru.GetOrPrepare("0", TTL)
			}
		})
	})
	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := strconv.Itoa(i)
			lru.GetOrPrepare(key, TTL)
			lru.Update(key, proto.Message{})
		}
	})
}
