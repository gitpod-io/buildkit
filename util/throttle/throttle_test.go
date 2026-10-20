package throttle

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottle(t *testing.T) {
	t.Parallel()

	var i atomic.Int64
	f := func() {
		i.Add(1)
	}

	f = Throttle(50*time.Millisecond, f)

	f()
	f()

	require.Equal(t, int64(0), i.Load())

	// test that i is never incremented twice and at least once in next 600ms
	retries := 0
	for {
		require.Less(t, retries, 10)
		time.Sleep(60 * time.Millisecond)
		v := i.Load()
		require.LessOrEqual(t, v, int64(1))
		if v == 1 {
			break
		}
		retries++
	}

	require.Equal(t, int64(1), i.Load())

	f()

	retries = 0
	for {
		require.Less(t, retries, 10)
		time.Sleep(60 * time.Millisecond)
		v := i.Load()
		if v == 2 {
			break
		}
		retries++
	}
}

func TestAfter(t *testing.T) {
	t.Parallel()

	var i atomic.Int64
	f := func() {
		i.Add(1)
	}

	f = After(100*time.Millisecond, f)

	f()

	time.Sleep(10 * time.Millisecond)
	require.Equal(t, int64(1), i.Load())
	f()
	time.Sleep(10 * time.Millisecond)
	require.Equal(t, int64(1), i.Load())

	time.Sleep(200 * time.Millisecond)
	require.Equal(t, int64(2), i.Load())
}
