package crawler

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiterMaxTakesExceed(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()

	deferTime := 100 * time.Millisecond

	limiter := NewLimiter(LimiterConfig{DeferTime: deferTime, MaxTakes: 4})
	numbers, subs := mockProducer(10, limiter)
	limiter.Wait()

	assert.Equal(t, []int{0, 1, 2, 3}, numbers)

	for idx := 0; idx < len(subs)-1; idx++ {
		diff := subs[idx+1] - subs[idx]
		assert.Greater(t, diff, deferTime-5*time.Millisecond)
		assert.Less(t, diff, deferTime+5*time.Millisecond)
	}

	time.Sleep(deferTime)
	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

func TestLimiterMaxTakesNotExceed(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()

	limiter := NewLimiter(LimiterConfig{MaxTakes: 4})
	numbers, _ := mockProducer(3, limiter)
	limiter.Wait()

	assert.Equal(t, []int{0, 1, 2}, numbers)
	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

/* Private stuffs */

func mockProducer(amount int, limiter Limiter) ([]int, []time.Duration) {
	numbersCh := make(chan []int)
	subsCh := make(chan []time.Duration)

	go func() {
		start := time.Now()
		numbers := []int{}
		subs := []time.Duration{}

		for num := 0; num < amount; num++ {
			ok := limiter.Take()
			now := time.Now()

			if !ok {
				break
			}

			numbers = append(numbers, num)
			subs = append(subs, now.Sub(start))
			limiter.Done(1)
		}

		numbersCh <- numbers
		subsCh <- subs
	}()

	numbers := <-numbersCh
	subs := <-subsCh
	return numbers, subs
}
