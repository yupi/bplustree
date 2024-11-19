package bplustree_test

import (
	"testing"

	"golang.org/x/exp/slices"

	"bplustree"
)

func TestIteratorEmpty(t *testing.T) {
	tree := bplustree.NewBPlusTree[int64, string](2)
	it := tree.Begin()
	if !it.Done() {
		t.Fail()
	}
	it = tree.End()
	if !it.Done() {
		t.Fail()
	}
}

func TestIteratorClone(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")

	it1 := tree.LowerBound(4)
	it2 := it1.Clone()
	if (it1.Key() != it2.Key()) || (it1.Value() != it2.Value()) {
		t.Fail()
	}
}

func TestIteratorNextPrev(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	it := tree.End()
	if it.Done() {
		t.Fail()
	}
	if it.Next() || !it.Done() {
		t.Fail()
	}
	if it.Next() {
		t.Fail()
	}

	it = tree.Begin()
	if it.Done() {
		t.Fail()
	}
	if it.Prev() || !it.Done() {
		t.Fail()
	}
	if it.Prev() {
		t.Fail()
	}
}

func TestIteratorRange(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(2, "b")
	tree.Insert(3, "c")
	tree.Insert(4, "d")
	tree.Insert(5, "e")

	var keys []int
	var values []string

	tree.Delete(3)
	for v := range tree.Begin().Seq(true) {
		values = append(values, v)
	}
	tree.Delete(4)
	for v := range tree.End().Seq(false) {
		values = append(values, v)
	}
	if !slices.Equal(values, []string{"a", "b", "d", "e", "e", "b", "a"}) {
		t.Errorf("expected [a b d e e b a] got %v", values)
	}

	tree.Delete(1)
	values = []string{}
	for k, v := range tree.Begin().Seq2(true) {
		keys = append(keys, k)
		values = append(values, v)
	}
	for k, v := range tree.End().Seq2(false) {
		keys = append(keys, k)
		values = append(values, v)
	}
	if !slices.Equal(keys, []int{2, 5, 5, 2}) {
		t.Errorf("expected [2 5 5 2] got %v", keys)
	}
	if !slices.Equal(values, []string{"b", "e", "e", "b"}) {
		t.Errorf("expected [b e e b] got %v", values)
	}

	tree.Insert(3, "f")
	values = []string{}
	for v := range tree.Begin().Seq(true) {
		if v == "e" {
			break
		}
		values = append(values, v)
	}
	if !slices.Equal(values, []string{"b", "f"}) {
		t.Errorf("expected [b f] got %v", values)
	}

	values = []string{}
	for k, v := range tree.End().Seq2(false) {
		if k == 2 {
			break
		}
		values = append(values, v)
	}
	if !slices.Equal(values, []string{"e", "f"}) {
		t.Errorf("expected [e f] got %v", values)
	}
}

func TestInsertDeleteLen(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(2, "b")
	tree.Insert(3, "c")
	tree.Insert(4, "d")
	tree.Insert(5, "e")
	tree.Insert(6, "f")
	tree.Insert(7, "g")
	tree.Insert(8, "h")
	tree.Insert(9, "i")
	tree.Insert(10, "j")
	tree.Insert(11, "k")
	tree.Insert(12, "l")
	tree.Insert(13, "m")
	tree.Insert(14, "n")
	tree.Insert(15, "o")
	tree.Insert(16, "p")
	tree.Insert(17, "q")
	tree.Insert(18, "r")
	if !tree.Delete(10) {
		t.Error("tree.Delete() returns false instead of true")
	}
	if tree.Delete(10) {
		t.Error("tree.Delete() returns true instead of false")
	}

	if tree.Len() != 17 {
		t.Errorf("expected tree.Len() == 17 got %d", tree.Len())
	}
}

func TestLowerBound(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
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

func TestUpperBound(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	it := tree.UpperBound(4)
	if it.Value() != "x" {
		t.Errorf("expected 'x', got '%s'", it.Value())
	}
	it = tree.UpperBound(3)
	if it.Value() != "x" {
		t.Errorf("expected 'x', got '%s'", it.Value())
	}
	it = tree.UpperBound(0)
	if !it.Done() {
		t.Error("iterator not done")
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

func TestGraphviz(t *testing.T) {
	tree := bplustree.NewBPlusTree[int, string](2)
	tree.Insert(1, "a")
	tree.Insert(3, "b")
	tree.Insert(7, "c")
	tree.Insert(8, "d")
	tree.Insert(5, "e")
	tree.Insert(3, "x")

	err := bplustree.Graphviz(tree, "bplustree.svg")
	if err != nil {
		t.Error(err)
	}
}
