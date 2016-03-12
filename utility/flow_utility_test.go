package utility

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultipleConfirm(t *testing.T) {
	counter := 0

	// case for isTenable holds true
	isTenable := func() bool {
		counter++

		return true
	}
	MultipleConfirm(3, isTenable, time.Second)

	assert.Equal(t, 3, counter)

	// case for isTenable returns false
	isTenable = func() bool {
		return false
	}
	assert.False(t, MultipleConfirm(3, isTenable, time.Second))
}

func TestRunUntilSuccess(t *testing.T) {

	// case for failure
	counter := 0
	f := func() bool {
		counter++

		if counter > 5 {
			return true
		} else {
			return false
		}
	}

	assert.False(t, RunUntilSuccess(3, f, time.Second))

	// case for success
	counter = 0
	f = func() bool {
		counter++

		if counter > 2 {
			return true
		} else {
			return false
		}
	}

	assert.True(t, RunUntilSuccess(3, f, time.Second))
}
