package vhll

import (
	"strconv"
	"testing"
)

func TestVHLL(t *testing.T) {
	vhll, _ := NewForLog2m(20)
	for i := uint(0); i <= 1000000; i++ {
		for j := uint(1); j <= 5; j++ {
			if i%j == 0 {
				id := []byte(strconv.Itoa(int(j)))
				vhll.Add([]byte(id), []byte(strconv.Itoa(int(i))))
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
		card := float64(vhll.GetCardinality(id))
		offset := 100 * (card - float64(expected[j])) / float64(expected[j])
		if offset > 4 || offset < -4 {
			t.Error("Expected error < 4 percent, got", offset,
				", expected count for", j, "=", expected[j], "got", card)
		}
	}
	//fmt.Println(vhll.GetTotalCardinality())
	/*
		totalOffset := 100 * (float64(vhll.GetTotalCardinality()) - float64(10000000)) / float64(10000000)
		if totalOffset > 13 || totalOffset < -13 {
			t.Error("Expected error < 13 percent, got", totalOffset, vhll.GetTotalCardinality())
		}
	*/
}
