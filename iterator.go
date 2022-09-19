package zset

type RangeIterator struct {
	node            *node
	start, end, cur int
	reverse         bool
}

func (r *RangeIterator) Len() int {
	return r.end - r.start + 1
}

func (r *RangeIterator) Valid() bool {
	return r.cur <= r.end
}

func (r *RangeIterator) Next() {
	if r.reverse {
		r.node = r.node.backward
	} else {
		r.node = r.node.level[0].forward
	}
	r.cur++
}

func (r *RangeIterator) Item() Item {
	return r.node.item
}

func (r *RangeIterator) Key() string {
	return r.node.key
}

func (r *RangeIterator) Rank() int {
	return r.cur + 1
}
