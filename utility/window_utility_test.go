package utility

import (
	"log"
	"testing"
)

func TestWindow(t *testing.T) {
	w := NewWindow(3, 1)

	w.PushBack("1001")
	ws := w.Slice()
	log.Printf("w slice: %v", ws)

	w.PushBack("2001")
	ws = w.Slice()
	log.Printf("w slice: %v", ws)

	w.PushBack("channl1")
	ws = w.Slice()
	log.Printf("w slice: %v", ws)

	w.PushBack("channel2")
	ws = w.Slice()
	log.Printf("w slice: %v", ws)

	w.PushBack("3001")
	ws = w.Slice()
	log.Printf("w slice: %v", ws)
}
