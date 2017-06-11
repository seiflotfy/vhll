package vhll

import (
	"strconv"
	"testing"
)

func TestVHLL(t *testing.T) {

	vhll, _ := NewVHLL(18, 12)
	for i := uint(0); i <= 1000000; i++ {
		for j := uint(1); j <= 5; j++ {
			if i%j == 0 {
				id := []byte(strconv.Itoa(int(j)))
				vhll.Insert([]byte(id), []byte(strconv.Itoa(int(i))))
			}
		}
	}

	expected := make(map[uint]uint)
	expected[1] = 1000000
	expected[2] = 500000
	expected[3] = 333333
	expected[4] = 250000
	expected[5] = 200000

	for j := uint(1); j <= 5; j++ {
		id := []byte(strconv.Itoa(int(j)))
		card := float64(vhll.Estimate(id))
		p4 := float64(4 * card / 100)
		if float64(card) > float64(expected[j])+p4 || float64(card) < float64(expected[j])-p4 {
			t.Error("Expected error < 4 percent, got count for", j, "=", expected[j], "got", card)
		}
	}
}
