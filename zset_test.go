package zset

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func init() {
	seed := time.Now().Unix()
	rand.Seed(seed)
}

// perm returns a random permutation of n Int items in the range [0, n).
func perm(n int) (out []Int) {
	out = make([]Int, 0, n)
	for _, v := range rand.Perm(n) {
		out = append(out, Int(v))
	}
	return
}

// rang returns an ordered list of Int items in the range [0, n).
func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, Int(i))
	}
	return
}

func TestZSetRank(t *testing.T) {
	const listSize = 10000
	zs := New()
	for i := 0; i < 10; i++ {
		for _, v := range perm(listSize) {
			zs.Add(v)
		}
		for _, v := range perm(listSize) {
			if zs.Rank(v.Key(), false) != int(v)+1 {
				t.Error("rank error")
			}
			if zs.Rank(v.Key(), true) != int(listSize-v) {
				t.Error("rank error")
			}
		}

		if r := zs.Range(0, 1, false); !reflect.DeepEqual(r, rang(2)) {
			t.Error("range error")
		}

		if r := zs.Range(0, 1, true); r[0] != Int(listSize-1) || r[1] != Int(listSize-2) {
			t.Error("range error")
		}

		for i := 0; i < listSize/2; i++ {
			zs.Delete(Int(i).Key())
		}
		for i := listSize + 1; i < listSize; i++ {
			if r := zs.Rank(Int(i).Key(), false); r != i-listSize/2 {
				t.Error("rank failed")
			}
		}
	}
}

const benchmarkListSize = 10000

func BenchmarkAdd(b *testing.B) {
	zs := New()
	items := perm(benchmarkListSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zs.Add(items[i%benchmarkListSize])
	}
}
