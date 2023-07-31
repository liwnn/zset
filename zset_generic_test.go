//go:build go1.18

package zset

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

type TestRank struct {
	member string
	score  int
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
func rang(n int) (out []TestRank) {
	for i := 0; i < n; i++ {
		out = append(out, TestRank{
			member: strconv.Itoa(i),
			score:  i,
		})
	}
	return
}

func revrang(n int, count int) (out []TestRank) {
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
	zs := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for i := 0; i < 10; i++ {
		for _, v := range perm(listSize) {
			zs.Add(v.member, v)
		}
		for _, v := range perm(listSize) {
			if zs.Rank(v.member, false) != v.score+1 {
				t.Error("rank error")
			}
			if zs.Rank(v.member, true) != listSize-v.score {
				t.Error("rank error")
			}
		}

		var r []TestRank
		zs.Range(0, 1, false, func(item TestRank, _ int) bool {
			r = append(r, item)
			return true
		})
		if !reflect.DeepEqual(r, rang(2)) {
			t.Error("range error")
		}

		r = r[:0]
		zs.RangeByScore(func(i TestRank) bool {
			return i.score >= 0
		}, func(i TestRank) bool {
			return i.score <= 1
		}, false, func(item TestRank, rank int) bool {
			r = append(r, item)
			return true
		})
		if !reflect.DeepEqual(r, rang(2)) {
			t.Error("RangeItem error", r, rang(2))
		}

		r = r[:0]
		zs.Range(0, 1, true, func(item TestRank, _ int) bool {
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
	zs := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	zs.RangeByScore(nil, nil, false, func(i TestRank, rank int) bool {
		return true
	})

	for _, i := range perm(10) {
		zs.Add(i.member, i)
	}

	var r []TestRank
	zs.RangeByScore(nil, nil, false, func(i TestRank, rank int) bool {
		r = append(r, i)
		return true
	})
	if !reflect.DeepEqual(r, rang(10)) {
		t.Error("RangeItem error", r, rang(10))
	}

	r = r[:0]
	zs.RangeByScore(func(i TestRank) bool {
		return i.score >= 3
	}, func(i TestRank) bool {
		return i.score <= 5
	}, false, func(i TestRank, rank int) bool {
		r = append(r, i)
		return true
	})
	var expect []TestRank
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
	zs.RangeByScore(func(i TestRank) bool {
		return i.score >= 3
	}, func(i TestRank) bool {
		return i.score <= 5
	}, true, func(i TestRank, rank int) bool {
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
		tr := New[string, TestRank](func(a, b TestRank) bool {
			return a.score < b.score
		})
		for _, item := range insertP {
			tr.Add(item.member, item)
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
		tr := New[string, TestRank](func(a, b TestRank) bool {
			return a.score < b.score
		})
		for _, item := range insertP {
			tr.Add(item.member, item)
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
		tr := New[string, TestRank](func(a, b TestRank) bool {
			return a.score < b.score
		})
		for _, item := range insertP {
			tr.Add(item.member, item)
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
	tr := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for _, item := range insertP {
		tr.Add(item.member, item)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Remove(insertP[i%benchmarkListSize].member)
		item := insertP[i%benchmarkListSize]
		tr.Add(item.member, item)
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
		tr := New[string, TestRank](func(a, b TestRank) bool {
			return a.score < b.score
		})
		for _, item := range insertP {
			tr.Add(item.member, item)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Remove(item.member)
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
	tr := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for _, item := range insertP {
		tr.Add(item.member, item)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Rank(insertP[i%benchmarkListSize].member, true)
	}
}

func BenchmarkRange(b *testing.B) {
	insertP := perm(benchmarkListSize)
	tr := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for _, item := range insertP {
		tr.Add(item.member, item)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Range(0, 100, true, func(i TestRank, rank int) bool {
			return true
		})
	}
}

func BenchmarkRangeIterator(b *testing.B) {
	insertP := perm(benchmarkListSize)
	tr := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for _, item := range insertP {
		tr.Add(item.member, item)
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
	tr := New[string, TestRank](func(a, b TestRank) bool {
		return a.score < b.score
	})
	for _, item := range insertP {
		tr.Add(item.member, item)
	}
	minScore, maxScore := 0, 100
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.RangeByScore(func(i TestRank) bool {
			return i.score >= minScore
		}, func(i TestRank) bool {
			return i.score <= maxScore
		}, true, func(i TestRank, rank int) bool {
			return true
		})
	}
}
