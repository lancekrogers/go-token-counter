package spm

// priorityQueue is a generic priority queue backed by a binary heap.
type priorityQueue[T any] struct {
	cmp   func(a, b T) int
	items []T
}

func newPriorityQueue[T any](sizeHint int, cmp func(a, b T) int) *priorityQueue[T] {
	return &priorityQueue[T]{cmp: cmp, items: make([]T, 1, max(1, sizeHint+1))}
}

func (pq *priorityQueue[T]) Len() int {
	return len(pq.items) - 1
}

func (pq *priorityQueue[T]) Insert(elem T) {
	pq.items = append(pq.items, elem)
	pq.siftup(len(pq.items) - 1)
}

func (pq *priorityQueue[T]) PopMax() T {
	if len(pq.items) < 2 {
		panic("popping from empty priority queue")
	}
	maxItem := pq.items[1]
	pq.items[1] = pq.items[len(pq.items)-1]
	pq.items = pq.items[:len(pq.items)-1]
	pq.siftdown(1)
	return maxItem
}

func (pq *priorityQueue[T]) RemoveFunc(rm func(T) bool) {
	i := 1
	for ; i < len(pq.items); i++ {
		if rm(pq.items[i]) {
			break
		}
	}
	if i == len(pq.items) {
		return
	}
	for j := i + 1; j < len(pq.items); j++ {
		if v := pq.items[j]; !rm(v) {
			pq.items[i] = v
			i++
		}
	}
	clear(pq.items[i:])
	pq.items = pq.items[:i]
	pq.rebuildHeap()
}

func (pq *priorityQueue[T]) rebuildHeap() {
	for i := len(pq.items) / 2; i >= 1; i-- {
		pq.siftdown(i)
	}
}

func (pq *priorityQueue[T]) siftup(n int) {
	i := n
	for {
		if i == 1 {
			return
		}
		p := i / 2
		if pq.cmp(pq.items[p], pq.items[i]) >= 0 {
			return
		}
		pq.items[i], pq.items[p] = pq.items[p], pq.items[i]
		i = p
	}
}

func (pq *priorityQueue[T]) siftdown(i int) {
	for {
		c := 2 * i
		if c >= len(pq.items) {
			return
		}
		maxChild := c
		if c+1 < len(pq.items) {
			if pq.cmp(pq.items[c+1], pq.items[c]) > 0 {
				maxChild = c + 1
			}
		}
		if pq.cmp(pq.items[i], pq.items[maxChild]) >= 0 {
			return
		}
		pq.items[i], pq.items[maxChild] = pq.items[maxChild], pq.items[i]
		i = maxChild
	}
}
