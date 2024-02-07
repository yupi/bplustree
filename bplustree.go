package bplustree

import (
	"fmt"
	"io"
	"log"

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
	tree *BPlusTree[K, V]
	node *node[K, V]
	idx  int
	done bool
}

func (it *Iterator[K, V]) Key() K {
	return it.node.keys[it.idx]
}

func (it *Iterator[K, V]) Value() V {
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

func (it *Iterator[K, V]) Seek(key K) {
	*it = it.tree.LowerBound(key)
}

func (it *Iterator[K, V]) Done() bool {
	return it.done
}

func (it *Iterator[K, V]) Clone() Iterator[K, V] {
	return Iterator[K, V]{
		tree: it.tree,
		node: it.node,
		idx:  it.idx,
		done: it.done,
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
	insertAt(&parent.keys, i, key)
	insertAt(&parent.children, i+1, right)

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
	for i, k := range node.keys {
		if key == k {
			return node.values[i], true
		}
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
	insertAt(&node.keys, i, key)
	insertAt(&node.values, i, value)
}

/*
func (tree *BPlusTree[K, V]) InsertMulti(key K, value V) {
	node := tree.findLeaf(key)
	tree.insertLeafMulti(node, key, value)
	if len(node.keys) > 2*tree.order {
		tree.splitLeaf(node)
	}
}

func (tree *BPlusTree[K, V]) insertLeafMulti(node *node[K, V], key K, value V) {
	i := indexLess(node.keys, key)
	insertAt(&node.keys, i, key)
	insertAt(&node.values, i, value)
}
*/

func (tree *BPlusTree[K, V]) Delete(key K) bool {
	node := tree.findLeaf(key)
	for i, k := range node.keys {
		if key == k {
			tree.len--
			removeAt(&node.keys, i)
			removeAt(&node.values, i)
			return true
		}
	}
	return false
}

func (tree *BPlusTree[K, V]) LowerBound(key K) Iterator[K, V] {
	node := tree.findLeaf(key)
	for i, k := range node.keys {
		if k >= key {
			return Iterator[K, V]{tree: tree, node: node, idx: i}
		}
	}
	return Iterator[K, V]{done: true}
}

func (tree *BPlusTree[K, V]) IteratorBegin() Iterator[K, V] {
	it := Iterator[K, V]{
		tree: tree,
		node: tree.begin,
	}
	if len(it.node.keys) == 0 {
		it.next()
	}
	return it
}

func (tree *BPlusTree[K, V]) IteratorEnd() Iterator[K, V] {
	it := Iterator[K, V]{
		tree: tree,
		node: tree.end,
		idx:  len(tree.end.keys) - 1,
	}
	if len(it.node.keys) == 0 {
		it.prev()
	}
	return it
}

func insertAt[T any](s *[]T, index int, item T) {
	var zero T
	*s = append(*s, zero)
	if index < len(*s) {
		_ = copy((*s)[index+1:], (*s)[index:])
	}
	(*s)[index] = item
}

func removeAt[T any](s *[]T, index int) {
	*s = append((*s)[:index], (*s)[index+1:]...)
}

func indexLess[V constraints.Ordered](s []V, v V) int {
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

func (tree *BPlusTree[K, V]) Dump() {
	node := tree.root
	dumpNode(node)
}

func dumpNode[K constraints.Ordered, V any](n *node[K, V]) {
	log.Printf("node %p", n)
	log.Printf("parent %p", n.parent)
	log.Printf("keys: %v", n.keys)
	log.Printf("children: %v", n.children)
	log.Printf("values: %+v", n.values)
	log.Printf("prev: %p", n.prev)
	log.Printf("next: %p", n.next)
	log.Println()

	for _, c := range n.children {
		dumpNode(c)
	}
}

func DumpLeaves[K constraints.Ordered, V any](it *Iterator[K, V], w io.Writer) {
	for ; !it.Done(); it.Next() {
		var k interface{} = it.Key()
		if _, ok := k.(fmt.Stringer); ok {
			_, _ = io.WriteString(w, fmt.Sprintf(" '%s'", k))
		} else {
			_, _ = io.WriteString(w, fmt.Sprintf(" '%v'", k))
		}
		//_, _ = io.WriteString(w, fmt.Sprintf("%v\n", it.Value()))
	}
}
