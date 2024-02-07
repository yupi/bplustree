package bplustree_test

import (
	"testing"

	"golang.org/x/exp/slices"

	"bplustree"
)

func TestEmptyTreeIterator(t *testing.T) {
	tree := bplustree.NewBPlusTree[int64, string](2)
	it := tree.IteratorBegin()
	if !it.Done() {
		t.Fail()
	}
	it = tree.IteratorEnd()
	if !it.Done() {
		t.Fail()
	}
}

func TestIterators(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](1)
	tree.Insert(1, "a")
	tree.Insert(2, "b")
	tree.Insert(3, "c")
	tree.Insert(4, "d")
	tree.Insert(5, "e")

	tree.Delete(3)
	tree.Delete(4)
	var keys []int
	for it := tree.IteratorBegin(); !it.Done(); it.Next() {
		keys = append(keys, it.Key())
	}
	for it := tree.IteratorEnd(); !it.Done(); it.Prev() {
		keys = append(keys, it.Key())
	}
	if !slices.Equal(keys, []int{1, 2, 5, 5, 2, 1}) {
		t.Errorf("expected [1 2 5 5 2 1] got %v", keys)
	}

	tree.Delete(1)
	if tree.Delete(1) {
		t.Fail()
	}
	keys = []int{}
	for it := tree.IteratorBegin(); !it.Done(); it.Next() {
		keys = append(keys, it.Key())
	}
	for it := tree.IteratorEnd(); !it.Done(); it.Prev() {
		keys = append(keys, it.Key())
	}
	if !slices.Equal(keys, []int{2, 5, 5, 2}) {
		t.Errorf("expected [2 5 5 2] got %v", keys)
	}

	tree.Delete(2)
	keys = []int{}
	for it := tree.IteratorBegin(); !it.Done(); it.Next() {
		keys = append(keys, it.Key())
	}
	for it := tree.IteratorEnd(); !it.Done(); it.Prev() {
		keys = append(keys, it.Key())
	}
	if !slices.Equal(keys, []int{5, 5}) {
		t.Errorf("expected [5 5] got %v", keys)
	}
}

func TestLowerBound(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](1)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	it := tree.LowerBound(4)
	if it.Value() != "e" {
		t.Errorf("expected 'e', got '%s'", it.Value())
	}
	it = tree.LowerBound(3)
	if it.Value() != "x" {
		t.Errorf("expected 'x', got '%s'", it.Value())
	}
	it = tree.LowerBound(9)
	if !it.Done() {
		t.Error("iterator not done")
	}
}

func TestNextPrev(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	it := tree.IteratorEnd()
	if it.Done() {
		t.Fail()
	}
	if it.Next() || !it.Done() {
		t.Fail()
	}
	if it.Next() {
		t.Fail()
	}

	it = tree.IteratorBegin()
	if it.Done() {
		t.Fail()
	}
	if it.Prev() || !it.Done() {
		t.Fail()
	}
	if it.Prev() {
		t.Fail()
	}

	values := []string{}
	for it := tree.IteratorBegin(); !it.Done(); it.Next() {
		values = append(values, it.Value())
	}
	if !slices.Equal(values, []string{"a", "x", "e", "c", "d"}) {
		t.Errorf("expected [a x e c d] got %v", values)
	}

	values = []string{}
	for it := tree.IteratorEnd(); !it.Done(); it.Prev() {
		values = append(values, it.Value())
	}
	if !slices.Equal(values, []string{"d", "c", "e", "x", "a"}) {
		t.Errorf("expected [d c e x a] got %v", values)
	}
}

func TestSearch(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	v, ok := tree.Search(7)
	if v != "c" || !ok {
		t.Fail()
	}
	v, ok = tree.Search(4)
	if v != "" || ok {
		t.Fail()
	}
	v, ok = tree.Search(3)
	if v != "x" || !ok {
		t.Fail()
	}
}

func TestIteratorSeek(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	it := tree.IteratorBegin()
	it.Seek(6)
	if it.Key() != 7 && it.Value() != "c" {
		t.Fail()
	}

	it.Seek(3)
	if it.Key() != 3 && it.Value() != "x" {
		t.Fail()
	}
}

/*
func TestMulti(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](1)
	tree.InsertMulti(1, "a")
	tree.InsertMulti(3, "b")
	tree.InsertMulti(5, "e")
	tree.InsertMulti(3, "x")
	tree.InsertMulti(3, "y")
	tree.Dump()
	values := []string{}
	for it := tree.IteratorBegin(); !it.Done(); it.Next() {
		values = append(values, it.Value())
	}
	if !slices.Equal(values, []string{"a", "b", "x", "y", "e"}) {
		t.Errorf("expected [a b x y e] got %v", values)
	}

	v, ok := tree.Search(5)
	if !ok || v != "e" {
		t.Fail()
	}
	v, ok = tree.Search(3) // FIXME
	if !ok || v != "y" {
		t.Fail()
	}

	it := tree.LowerBound(3)
	t.Error(it.Value())
}
*/
