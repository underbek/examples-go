package concurrency

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"unsafe"
)

type mutexSlice []*sync.Mutex

func (m mutexSlice) Len() int {
	return len(m)
}

func (m mutexSlice) Less(i, j int) bool {
	return uintptr(unsafe.Pointer(m[i])) < uintptr(unsafe.Pointer(m[j]))
}

func (m mutexSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type Guard struct {
	mtxSlice mutexSlice
}

func NewGuard(mutexes ...*sync.Mutex) *Guard {
	mtxSlice := mutexSlice(mutexes)
	sort.Sort(mtxSlice)

	return &Guard{
		mtxSlice: mtxSlice,
	}
}

func (g *Guard) Lock() {
	for _, mtx := range g.mtxSlice {
		mtx.Lock()
	}
}

func (g *Guard) Unlock() {
	for _, mtx := range g.mtxSlice {
		mtx.Unlock()
	}
}

type User struct {
	amount uint
	mtx    sync.Mutex
}

func transaction(user1, user2 *User, sum uint) error {

	guard := NewGuard(&user1.mtx, &user2.mtx)
	guard.Lock()
	defer guard.Unlock()

	if user1.amount < sum {
		return errors.New("not enough money")
	}

	user1.amount -= sum
	user2.amount += sum

	return nil
}

func TestBalance(t *testing.T) {
	user1 := &User{amount: 100}

	user2 := &User{amount: 10}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		transaction(user1, user2, 10)
	}()

	go func() {
		defer wg.Done()
		transaction(user2, user1, 10)
	}()

	wg.Wait()

	assert.Equal(t, 100, int(user1.amount))
	assert.Equal(t, 10, int(user2.amount))
}
