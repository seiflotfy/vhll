package vhll

import (
	"strconv"
	"testing"
)

func TestVHLL(t *testing.T) {
	vhll, _ := NewForLog2m(24)
	for i := uint(0); i <= 2000000; i++ {
		for j := uint(1); j <= 5; j++ {
			if i%j == 0 {
				id := []byte(strconv.Itoa(int(j)))
				vhll.Add([]byte(id), []byte(strconv.Itoa(int(i))))
			}
		}
	}

	expected := make(map[uint]uint)
	expected[1] = 1000000 * 2
	expected[2] = 500000 * 2
	expected[3] = 333333 * 2
	expected[4] = 250000 * 2
	expected[5] = 200000 * 2

	for j := uint(1); j <= 5; j++ {
		id := []byte(strconv.Itoa(int(j)))
		card := float64(vhll.GetCardinality(id))
		offset := 100 * (card - float64(expected[j])) / float64(expected[j])
		if offset > 13 || offset < -13 {
			t.Error("Expected error < 13 percent, got", offset,
				", expected count for", j, "=", expected[j], "got", card)
		}
	}

	totalOffset := 100 * (float64(vhll.GetTotalCardinality()) - float64(10000000)) / float64(10000000)
	if totalOffset > 13 || totalOffset < -13 {
		t.Error("Expected error < 13 percent, got", totalOffset, vhll.GetTotalCardinality())
	}

}
