package zset

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func init() {
	seed := time.Now().Unix()
	rand.Seed(seed)
}

type TestRank struct {
	member string
	score  int
}

func (a TestRank) Key() string {
	return a.member
}

func (a TestRank) Less(than Item) bool {
	return a.score < than.(TestRank).score
}

// perm returns a random permutation of n Int items in the range [0, n).
func perm(n int) (out []TestRank) {
	out = make([]TestRank, 0, n)
	for _, v := range rand.Perm(n) {
		out = append(out, TestRank{
			member: strconv.Itoa(v),
			score:  v,
		})
	}
	return
}

// rang returns an ordered list of Int items in the range [0, n).
func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, TestRank{
			member: strconv.Itoa(i),
			score:  i,
		})
	}
	return
}

func revrang(n int, count int) (out []Item) {
	for i := n - 1; i >= n-count; i-- {
		out = append(out, TestRank{
			member: strconv.Itoa(i),
			score:  i,
		})
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
			if zs.Rank(v.Key(), false) != v.score+1 {
				t.Error("rank error")
			}
			if zs.Rank(v.Key(), true) != int(listSize-v.score) {
				t.Error("rank error")
			}
		}

		var r []Item
		zs.Range(0, 1, false, func(item Item, _ int) bool {
			r = append(r, item)
			return true
		})
		if !reflect.DeepEqual(r, rang(2)) {
			t.Error("range error")
		}

		r = r[:0]
		zs.RangeByScore(func(i Item) bool {
			return i.(TestRank).score >= 0
		}, func(i Item) bool {
			return i.(TestRank).score <= 1
		}, false, func(item Item, rank int) bool {
			r = append(r, item)
			return true
		})
		if !reflect.DeepEqual(r, rang(2)) {
			t.Error("RangeItem error", r, rang(2))
		}

		r = r[:0]
		zs.Range(0, 1, true, func(item Item, _ int) bool {
			r = append(r, item)
			return true
		})
		if !reflect.DeepEqual(r, revrang(listSize, 2)) {
			t.Error("range error")
		}

		for i := 0; i < listSize/2; i++ {
			zs.Remove(strconv.Itoa(i))
		}
		for i := listSize + 1; i < listSize; i++ {
			if r := zs.Rank(strconv.Itoa(i), false); r != i-listSize/2 {
				t.Error("rank failed")
			}
		}
	}
}

func TestRangeItem(t *testing.T) {
	zs := New()
	zs.RangeByScore(nil, nil, false, func(i Item, rank int) bool {
		return true
	})

	for _, i := range perm(10) {
		zs.Add(i)
	}

	var r []Item
	zs.RangeByScore(nil, nil, false, func(i Item, rank int) bool {
		r = append(r, i)
		return true
	})
	if !reflect.DeepEqual(r, rang(10)) {
		t.Error("RangeItem error", r, rang(10))
	}

	r = r[:0]
	zs.RangeByScore(func(i Item) bool {
		return i.(TestRank).score >= 3
	}, func(i Item) bool {
		return i.(TestRank).score <= 5
	}, false, func(i Item, rank int) bool {
		r = append(r, i)
		return true
	})
	var expect []Item
	for i := 3; i <= 5; i++ {
		expect = append(expect, TestRank{
			member: strconv.Itoa(i),
			score:  i,
		})
	}
	if !reflect.DeepEqual(r, expect) {
		t.Error("RangeItem error", r, expect)
	}

	r = r[:0]
	zs.RangeByScore(func(i Item) bool {
		return i.(TestRank).score >= 3
	}, func(i Item) bool {
		return i.(TestRank).score <= 5
	}, true, func(i Item, rank int) bool {
		r = append(r, i)
		return true
	})
	expect = expect[:0]
	for i := 5; i >= 3; i-- {
		expect = append(expect, TestRank{
			member: strconv.Itoa(i),
			score:  i,
		})
	}
	if !reflect.DeepEqual(r, expect) {
		t.Error("RangeItem error", r, expect)
	}
}

const benchmarkListSize = 10000

func BenchmarkAdd(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkListSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New()
		for _, item := range insertP {
			tr.Add(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkAddIncrease(b *testing.B) {
	b.StopTimer()
	insertP := rang(benchmarkListSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New()
		for _, item := range insertP {
			tr.Add(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkAddDecrease(b *testing.B) {
	b.StopTimer()
	insertP := revrang(benchmarkListSize, benchmarkListSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New()
		for _, item := range insertP {
			tr.Add(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkRemoveAdd(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkListSize)
	tr := New()
	for _, item := range insertP {
		tr.Add(item)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Remove(insertP[i%benchmarkListSize].Key())
		item := insertP[i%benchmarkListSize]
		tr.Add(item)
	}
}

func BenchmarkRemove(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkListSize)
	removeP := perm(benchmarkListSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New()
		for _, v := range insertP {
			tr.Add(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Remove(item.Key())
			i++
			if i >= b.N {
				return
			}
		}
		if tr.Length() > 0 {
			b.Error(tr.Length())
		}
	}
}

func BenchmarkRank(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkListSize)
	tr := New()
	for _, v := range insertP {
		tr.Add(v)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Rank(insertP[i%benchmarkListSize].Key(), true)
	}
}

func BenchmarkRange(b *testing.B) {
	insertP := perm(benchmarkListSize)
	tr := New()
	for _, item := range insertP {
		tr.Add(item)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Range(0, 100, true, func(i Item, rank int) bool {
			return true
		})
	}
}

func BenchmarkRangeIterator(b *testing.B) {
	insertP := perm(benchmarkListSize)
	tr := New()
	for _, item := range insertP {
		tr.Add(item)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := tr.RangeIterator(0, 100, true)
		for ; it.Valid(); it.Next() {
		}
	}
}

func BenchmarkRangeItem(b *testing.B) {
	insertP := perm(benchmarkListSize)
	tr := New()
	for _, item := range insertP {
		tr.Add(item)
	}
	minScore, maxScore := 0, 100
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.RangeByScore(func(i Item) bool {
			return i.(TestRank).score >= minScore
		}, func(i Item) bool {
			return i.(TestRank).score <= maxScore
		}, true, func(i Item, rank int) bool {
			return true
		})
	}
}
