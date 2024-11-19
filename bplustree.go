package bplustree

import (
	"iter"
	"slices"

	"golang.org/x/exp/constraints"
)

type node[K constraints.Ordered, V any] struct {
	keys     []K
	values   []V
	children []*node[K, V]
	parent   *node[K, V]
	next     *node[K, V]
	prev     *node[K, V]
}

func (n *node[K, V]) isLeaf() bool {
	return n.children == nil
}

type Iterator[K constraints.Ordered, V any] struct {
	node *node[K, V]
	idx  int
	done bool
}

func (it Iterator[K, V]) Key() K {
	return it.node.keys[it.idx]
}

func (it Iterator[K, V]) Value() V {
	return it.node.values[it.idx]
}

func (it *Iterator[K, V]) Next() bool {
	if it.done {
		return false
	}

	idx := it.idx + 1
	if idx < len(it.node.keys) {
		it.idx = idx
		return true
	}
	return it.next()
}

func (it *Iterator[K, V]) next() bool {
	for {
		if it.node.next == nil {
			it.done = true
			return false
		}
		it.node = it.node.next
		if len(it.node.keys) != 0 {
			it.idx = 0
			return true
		}
	}
}

func (it *Iterator[K, V]) Prev() bool {
	if it.done {
		return false
	}

	idx := it.idx - 1
	if idx >= 0 {
		it.idx = idx
		return true
	}
	return it.prev()
}

func (it *Iterator[K, V]) prev() bool {
	for {
		if it.node.prev == nil {
			it.done = true
			return false
		}
		it.node = it.node.prev
		if len(it.node.keys) != 0 {
			it.idx = len(it.node.keys) - 1
			return true
		}
	}
}

func (it Iterator[K, V]) Done() bool {
	return it.done
}

func (it Iterator[K, V]) Clone() Iterator[K, V] {
	return Iterator[K, V]{
		node: it.node,
		idx:  it.idx,
		done: it.done,
	}
}

func (it Iterator[K, V]) Seq(forward bool) iter.Seq[V] {
	advance := it.Next
	if !forward {
		advance = it.Prev
	}

	return func(yield func(V) bool) {
		for ; !it.done; advance() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

func (it Iterator[K, V]) Seq2(forward bool) iter.Seq2[K, V] {
	advance := it.Next
	if !forward {
		advance = it.Prev
	}

	return func(yield func(K, V) bool) {
		for ; !it.done; advance() {
			if !yield(it.Key(), it.Value()) {
				return
			}
		}
	}
}

type BPlusTree[K constraints.Ordered, V any] struct {
	order int
	root  *node[K, V]
	begin *node[K, V]
	end   *node[K, V]
	len   int
}

func NewBPlusTree[K constraints.Ordered, V any](order int) *BPlusTree[K, V] {
	tree := &BPlusTree[K, V]{
		order: order,
		root: &node[K, V]{
			keys: make([]K, 0, 2*order),
		},
	}
	tree.begin = tree.root
	tree.end = tree.root

	return tree
}

func (tree *BPlusTree[K, V]) findLeaf(key K) *node[K, V] {
	node := tree.root
	for !node.isLeaf() {
		i := indexLess(node.keys, key)
		node = node.children[i]
	}
	return node
}

func (tree *BPlusTree[K, V]) splitLeaf(n *node[K, V]) {
	newnode := &node[K, V]{
		parent: n.parent,
		keys:   make([]K, 0, 2*tree.order),
		values: make([]V, 0, 2*tree.order),
		next:   n.next,
		prev:   n,
	}
	newnode.keys = append(newnode.keys, n.keys[tree.order:]...)
	newnode.values = append(newnode.values, n.values[tree.order:]...)
	if newnode.next == nil {
		tree.end = newnode
	}

	n.next = newnode
	n.keys = n.keys[:tree.order]
	n.values = n.values[:tree.order]

	tree.splitNode(n.parent, n, newnode, newnode.keys[0])
}

func (tree *BPlusTree[K, V]) splitNode(parent, left, right *node[K, V], key K) {
	if parent == nil {
		newroot := &node[K, V]{
			keys:     make([]K, 0, 2*tree.order),
			children: make([]*node[K, V], 0, 2*tree.order+1),
			values:   nil,
		}
		newroot.keys = append(newroot.keys, key)
		newroot.children = append(newroot.children, left, right)
		left.parent = newroot
		right.parent = newroot
		tree.root = newroot
		return
	}

	i := indexLess(parent.keys, key)
	parent.keys = slices.Insert(parent.keys, i, key)
	parent.children = slices.Insert(parent.children, i+1, right)

	if len(parent.keys) <= 2*tree.order {
		return
	}

	newnode := &node[K, V]{
		parent:   parent,
		keys:     make([]K, 0, 2*tree.order),
		children: make([]*node[K, V], 0, 2*tree.order+1),
	}
	newnode.keys = append(newnode.keys, parent.keys[tree.order+1:]...)
	newnode.children = append(newnode.children, parent.children[tree.order+1:]...)
	left.parent = newnode
	right.parent = newnode

	key = parent.keys[tree.order]
	parent.keys = parent.keys[:tree.order]
	parent.children = parent.children[:tree.order+1]

	tree.splitNode(parent.parent, parent, newnode, key)
}

func (tree *BPlusTree[K, V]) Search(key K) (V, bool) {
	node := tree.findLeaf(key)
	idx, found := slices.BinarySearch(node.keys, key)
	if found {
		return node.values[idx], true
	}
	var zero V
	return zero, false
}

func (tree *BPlusTree[K, V]) Len() int {
	return tree.len
}

func (tree *BPlusTree[K, V]) Insert(key K, value V) {
	node := tree.findLeaf(key)
	tree.insertLeaf(node, key, value)
	if len(node.keys) > 2*tree.order {
		tree.splitLeaf(node)
	}
}

func (tree *BPlusTree[K, V]) insertLeaf(node *node[K, V], key K, value V) {
	i := indexLE(node.keys, key)
	if i < len(node.keys) && node.keys[i] == key {
		node.values[i] = value
		return
	}

	tree.len++
	node.keys = slices.Insert(node.keys, i, key)
	node.values = slices.Insert(node.values, i, value)
}

func (tree *BPlusTree[K, V]) Delete(key K) bool {
	node := tree.findLeaf(key)
	for i, k := range node.keys {
		if key == k {
			tree.len--
			node.keys = slices.Delete(node.keys, i, i+1)
			node.values = slices.Delete(node.values, i, i+1)
			return true
		}
	}
	return false
}

func (tree *BPlusTree[K, V]) LowerBound(key K) *Iterator[K, V] {
	node := tree.findLeaf(key)
	it := Iterator[K, V]{node: node}

	for i := range node.keys {
		if node.keys[i] >= key {
			return &Iterator[K, V]{node: node, idx: i}
		}
	}

	_ = it.next()
	return &it
}

func (tree *BPlusTree[K, V]) UpperBound(key K) *Iterator[K, V] {
	node := tree.findLeaf(key)
	it := Iterator[K, V]{node: node}

	for i := len(node.keys) - 1; i >= 0; i-- {
		if node.keys[i] <= key {
			return &Iterator[K, V]{node: node, idx: i}
		}
	}

	_ = it.prev()
	return &it
}

func (tree *BPlusTree[K, V]) Begin() Iterator[K, V] {
	it := Iterator[K, V]{node: tree.begin}
	if len(it.node.keys) == 0 {
		it.next()
	}
	return it
}

func (tree *BPlusTree[K, V]) End() Iterator[K, V] {
	it := Iterator[K, V]{
		node: tree.end,
		idx:  len(tree.end.keys) - 1,
	}
	if len(it.node.keys) == 0 {
		it.prev()
	}
	return it
}

func indexLess[S ~[]E, E constraints.Ordered](s S, v E) int {
	for i := range s {
		if v < s[i] {
			return i
		}
	}
	return len(s)
}

func indexLE[V constraints.Ordered](s []V, v V) int {
	for i := range s {
		if v <= s[i] {
			return i
		}
	}
	return len(s)
}
