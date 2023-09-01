//go:build go1.18

// Package zset implements sorted set similar to redis zset.
package zset

import (
	"math/rand"
	"time"
)

const (
	DefaultMaxLevel = 32   // (1/p)^MaxLevel >= maxNode
	DefaultP        = 0.25 // SkipList P = 1/4

	DefaultFreeListSize = 32
)

// ItemIterator allows callers of Range* to iterate of the zset.
// When this function returns false, iteration will stop.
type ItemIterator[T any] func(i T, rank int) bool

type skipListLevel[T any] struct {
	forward *node[T]
	span    int
}

// node is an element of a skip list
type node[T any] struct {
	item     T
	backward *node[T]
	level    []skipListLevel[T]
}

// FreeList represents a free list of set node.
type FreeList[T any] struct {
	freelist []*node[T]
}

// NewFreeList creates a new free list.
func NewFreeList[T any](size int) *FreeList[T] {
	return &FreeList[T]{freelist: make([]*node[T], 0, size)}
}

func (f *FreeList[T]) newNode(lvl int) (n *node[T]) {
	if len(f.freelist) == 0 {
		n = new(node[T])
		n.level = make([]skipListLevel[T], lvl)
		return
	}
	index := len(f.freelist) - 1
	n = f.freelist[index]
	f.freelist[index] = nil
	f.freelist = f.freelist[:index]

	if cap(n.level) < lvl {
		n.level = make([]skipListLevel[T], lvl)
	} else {
		n.level = n.level[:lvl]
	}
	return
}

func (f *FreeList[T]) freeNode(n *node[T]) (out bool) {
	// for gc
	var zero T
	n.item = zero
	for j := 0; j < len(n.level); j++ {
		n.level[j] = skipListLevel[T]{}
	}

	if len(f.freelist) < cap(f.freelist) {
		f.freelist = append(f.freelist, n)
		out = true
	}
	return
}

// skipList represents a skip list
type skipList[T any] struct {
	header, tail *node[T]
	length       int
	level        int // current level count
	maxLevel     int
	freelist     *FreeList[T]
	random       *rand.Rand
	less         LessFunc[T]
}

// newSkipList creates a skip list
func newSkipList[T any](maxLevel int, less LessFunc[T]) *skipList[T] {
	if maxLevel < DefaultMaxLevel {
		panic("maxLevel must < 32")
	}
	return &skipList[T]{
		level: 1,
		header: &node[T]{
			level: make([]skipListLevel[T], maxLevel),
		},
		maxLevel: maxLevel,
		freelist: NewFreeList[T](DefaultFreeListSize),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
		less:     less,
	}
}

// insert an item into the SkipList.
func (sl *skipList[T]) insert(item T) *node[T] {
	var update [DefaultMaxLevel]*node[T] // [0...list.maxLevel)
	var rank [DefaultMaxLevel]int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for y := x.level[i].forward; y != nil && sl.less(y.item, item); y = x.level[i].forward {
			rank[i] += x.level[i].span
			x = y
		}
		update[i] = x
	}

	lvl := sl.randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i].span = sl.length
		}
		sl.level = lvl
	}

	x = sl.freelist.newNode(lvl)
	x.item = item
	for i := 0; i < lvl; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := lvl; i < sl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward == nil {
		sl.tail = x
	} else {
		x.level[0].forward.backward = x
	}
	sl.length++
	return x
}

// delete element
func (sl *skipList[T]) delete(n *node[T]) (_ T) {
	var preAlloc [DefaultMaxLevel]*node[T] // [0...list.maxLevel)
	update := preAlloc[:sl.maxLevel]
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && sl.less(y.item, n.item); y = x.level[i].forward {
			x = y
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && !sl.less(n.item, x.item) {
		for i := 0; i < sl.level; i++ {
			if update[i].level[i].forward == x {
				update[i].level[i].span += x.level[i].span - 1
				update[i].level[i].forward = x.level[i].forward
			} else {
				update[i].level[i].span--
			}
		}
		for sl.level > 1 && sl.header.level[sl.level-1].forward == nil {
			sl.level--
		}
		if x.level[0].forward == nil {
			sl.tail = x.backward
		} else {
			x.level[0].forward.backward = x.backward
		}
		removeItem := x.item
		sl.freelist.freeNode(x)
		sl.length--
		return removeItem
	}
	return
}

func (sl *skipList[T]) updateItem(node *node[T], item T) bool {
	if (node.level[0].forward == nil || !sl.less(node.level[0].forward.item, item)) &&
		(node.backward == nil || !sl.less(item, node.backward.item)) {
		node.item = item
		return true
	}
	return false
}

// getRank find the rank for an element.
// Returns 0 when the element cannot be found, rank otherwise.
// Note that the rank is 1-based
func (sl *skipList[T]) getRank(item T) int {
	var rank int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && !sl.less(item, y.item); y = x.level[i].forward {
			rank += x.level[i].span
			x = y
		}
		if x != sl.header && !sl.less(x.item, item) {
			return rank
		}
	}
	return 0
}

func (sl *skipList[T]) randomLevel() int {
	lvl := 1
	for lvl < sl.maxLevel && float32(sl.random.Uint32()&0xFFFF) < DefaultP*0xFFFF {
		lvl++
	}
	return lvl
}

// Finds an element by its rank. The rank argument needs to be 1-based.
func (sl *skipList[T]) getNodeByRank(rank int) *node[T] {
	var traversed int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && traversed+x.level[i].span <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

func (sl *skipList[T]) getMinNode() *node[T] {
	return sl.header.level[0].forward
}

func (sl *skipList[T]) getMaxNode() *node[T] {
	return sl.tail
}

// return the first node greater and the node's 1-based rank.
func (sl *skipList[T]) findNext(greater func(i T) bool) (*node[T], int) {
	x := sl.header
	var rank int
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && !greater(y.item); y = x.level[i].forward {
			rank += x.level[i].span
			x = y
		}
	}
	return x.level[0].forward, rank + x.level[0].span
}

// return the first node less and the node's 1-based rank.
func (sl *skipList[T]) findPrev(less func(i T) bool) (*node[T], int) {
	var rank int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && less(y.item); y = x.level[i].forward {
			rank += x.level[i].span
			x = y
		}
	}
	return x, rank
}

// ZSet set
type ZSet[K comparable, T any] struct {
	dict map[K]*node[T]
	sl   *skipList[T]
}

// LessFunc determines how to order a type 'T'.  It should implement a strict
// ordering, and should return true if within that ordering, 'a' < 'b'.
type LessFunc[T any] func(a, b T) bool

// New creates a new ZSet.
func New[K comparable, T any](less LessFunc[T]) *ZSet[K, T] {
	return &ZSet[K, T]{
		dict: make(map[K]*node[T]),
		sl:   newSkipList[T](DefaultMaxLevel, less),
	}
}

// Add a new element or update the score of an existing element. If an item already
// exist, the removed item is returned. Otherwise, nil is returned.
func (zs *ZSet[K, T]) Add(key K, item T) (removeItem T) {
	if node := zs.dict[key]; node != nil {
		// if the node after update, would be still exactly at the same position,
		// we can just update item.
		if zs.sl.updateItem(node, item) {
			return
		}
		removeItem = zs.sl.delete(node)
	}
	zs.dict[key] = zs.sl.insert(item)
	return
}

// Remove the element 'ele' from the sorted set,
// return true if the element existed and was deleted, false otherwise
func (zs *ZSet[K, T]) Remove(key K) (removeItem T) {
	node := zs.dict[key]
	if node == nil {
		return
	}
	removeItem = zs.sl.delete(node)
	delete(zs.dict, key)
	return
}

// Rank return 1-based rank or 0 if not exist
func (zs *ZSet[K, T]) Rank(key K, reverse bool) int {
	node := zs.dict[key]
	if node != nil {
		rank := zs.sl.getRank(node.item)
		if rank > 0 {
			if reverse {
				return zs.sl.length - rank + 1
			}
			return rank
		}
	}
	return 0
}

func (zs *ZSet[K, T]) FindNext(iGreaterThan func(i T) bool) (v T, rank int) {
	n, rank := zs.sl.findNext(iGreaterThan)
	if n == nil {
		return
	}
	return n.item, rank
}

func (zs *ZSet[K, T]) FindPrev(iLessThan func(i T) bool) (v T, rank int) {
	n, rank := zs.sl.findPrev(iLessThan)
	if n == nil {
		return
	}
	return n.item, rank
}

// RangeByScore calls the iterator for every value within the range [min, max],
// until iterator return false. If min is nil, it represents negative infinity.
// If max is nil, it represents positive infinity.
func (zs *ZSet[K, T]) RangeByScore(min, max func(i T) bool, reverse bool, iterator ItemIterator[T]) {
	llen := zs.sl.length
	var minNode, maxNode *node[T]
	var minRank, maxRank int
	if min == nil {
		minNode = zs.sl.getMinNode()
		minRank = 1
	} else {
		minNode, minRank = zs.sl.findNext(min)
	}
	if minNode == nil {
		return
	}
	if max == nil {
		maxNode = zs.sl.getMaxNode()
		maxRank = llen
	} else {
		maxNode, maxRank = zs.sl.findPrev(max)
	}
	if maxNode == nil {
		return
	}
	if reverse {
		n := maxNode
		for i := maxRank; i >= minRank; i-- {
			if iterator(n.item, llen-i+1) {
				n = n.backward
			} else {
				break
			}
		}
	} else {
		n := minNode
		for i := minRank; i <= maxRank; i++ {
			if iterator(n.item, i) {
				n = n.level[0].forward
			} else {
				break
			}
		}
	}
}

// Range calls the iterator for every value with in index range [start, end],
// until iterator return false. The <start> and <stop> arguments represent
// zero-based indexes.
func (zs *ZSet[K, T]) Range(start, end int, reverse bool, iterator ItemIterator[T]) {
	llen := zs.sl.length
	if start < 0 {
		start = llen + start
	}
	if end < 0 {
		end = llen + end
	}
	if start < 0 {
		start = 0
	}
	if start > end || start >= llen {
		return
	}
	if end >= llen {
		end = llen - 1
	}

	rangeLen := end - start + 1
	if reverse {
		ln := zs.sl.getNodeByRank(llen - start)
		for i := 1; i <= rangeLen; i++ {
			if iterator(ln.item, start+i) {
				ln = ln.backward
			} else {
				break
			}
		}
	} else {
		ln := zs.sl.getNodeByRank(start + 1)
		for i := 1; i <= rangeLen; i++ {
			if iterator(ln.item, start+i) {
				ln = ln.level[0].forward
			} else {
				break
			}
		}
	}
}

type RangeIterator[T any] struct {
	node            *node[T]
	start, end, cur int
	reverse         bool
}

func (r *RangeIterator[T]) Len() int {
	return r.end - r.start + 1
}

func (r *RangeIterator[T]) Valid() bool {
	return r.cur <= r.end
}

func (r *RangeIterator[T]) Next() {
	if r.reverse {
		r.node = r.node.backward
	} else {
		r.node = r.node.level[0].forward
	}
	r.cur++
}

func (r *RangeIterator[T]) Item() T {
	return r.node.item
}

func (r *RangeIterator[T]) Rank() int {
	return r.cur + 1
}

// RangeIterator return iterator for visit elements in [start, end].
// It is slower than Range.
func (zs *ZSet[K, T]) RangeIterator(start, end int, reverse bool) RangeIterator[T] {
	llen := zs.sl.length
	if start < 0 {
		start = llen + start
	}
	if end < 0 {
		end = llen + end
	}
	if start < 0 {
		start = 0
	}

	if start > end || start >= llen {
		return RangeIterator[T]{end: -1}
	}

	if end >= llen {
		end = llen - 1
	}

	var n *node[T]
	if reverse {
		n = zs.sl.getNodeByRank(llen - start)
	} else {
		n = zs.sl.getNodeByRank(start + 1)
	}
	return RangeIterator[T]{
		start:   start,
		cur:     start,
		end:     end,
		node:    n,
		reverse: reverse,
	}
}

// Get return Item in dict.
func (zs *ZSet[K, T]) Get(key K) (item T, found bool) {
	if n, ok := zs.dict[key]; ok {
		return n.item, ok
	}
	return
}

// Length return the element count
func (zs *ZSet[K, T]) Length() int {
	return zs.sl.length
}
