package vhll

import (
	"errors"
	"math"

	"github.com/spaolacci/murmur3"
)

var mAlpha = []float64{
	0,
	0.44567926005415,
	1.2480639342271,
	2.8391255240079,
	6.0165231584811,
	12.369319965552,
	25.073991603109,
	50.482891762521,
	101.30047482549,
	202.93553337953,
	406.20559693552,
	812.74569741657,
	1625.8258887309,
	3251.9862249084,
	6504.3071471860,
	13008.949929672,
	26018.222470181,
	52036.684135280,
	104073.41696276,
	208139.24771523,
	416265.57100022,
	832478.53851627,
	1669443.2499579,
	3356902.8702907,
	6863377.8429508,
	11978069.823687,
	31333767.455026,
	52114301.457757,
	72080129.928986,
	68945006.880409,
	31538957.552704,
	3299942.4347441,
}

const alpha4SingleCounter float64 = 0.44567926005415

func log2m(rsd float64) uint {
	return uint(math.Log((1.106/rsd)*(1.106/rsd)) / math.Log(2))
}

func getVirtualEstimatorSize(physicalLog2m uint) uint {
	return physicalLog2m - 8
}

func getAlphaMM(log2m uint) float64 {
	m := math.Pow(2, float64(log2m))

	var alphaMM float64

	// See the paper.
	switch log2m {
	case 4:
		alphaMM = 0.673 * m * m
		break
	case 5:
		alphaMM = 0.697 * m * m
		break
	case 6:
		alphaMM = 0.709 * m * m
		break
	default:
		alphaMM = (0.7213 / (1 + 1.079/m)) * m * m
	}

	return alphaMM
}

/*
VirtualHyperLogLog ...
*/
type VirtualHyperLogLog struct {
	registers       *registerSet
	physicalLog2m   uint
	physicalM       uint
	physicalAlphaMM float64
	virtualLog2m    uint
	virtualM        uint
	virtualAlphaMM  float64
	virtualCa       float64
	//hll           hll.HyperLogLog
	murmur murmur3.Hash128
}

/*
NewForRsd creates a new VirtualHyperLogLog.
It takes rsd - the relative standard deviation for the counter.
smaller values create counters that require more space.
*/
func NewForRsd(rsd float64) (*VirtualHyperLogLog, error) {
	return NewForLog2m(log2m(rsd))
}

/*
NewForLog2m ...
*/
func NewForLog2m(log2m uint) (*VirtualHyperLogLog, error) {
	return New(log2m, newRegisterSet(uint(math.Pow(2, float64(log2m)))))
}

/*
New ...
*/
func New(physicalLog2m uint, registers *registerSet) (*VirtualHyperLogLog, error) {
	vhll := &VirtualHyperLogLog{}
	vhll.registers = registers
	vhll.physicalLog2m = physicalLog2m
	vhll.physicalM = uint(math.Pow(2, float64(physicalLog2m)))

	if physicalLog2m < 7 {
		return nil, errors.New("physicalLog2m needs to be >= 7")
	}

	vhll.virtualLog2m = getVirtualEstimatorSize(physicalLog2m)
	vhll.virtualAlphaMM = getAlphaMM(physicalLog2m)
	vhll.virtualM = uint(math.Pow(2, float64(vhll.virtualLog2m)))
	vhll.virtualCa = mAlpha[vhll.virtualLog2m]
	vhll.murmur = murmur3.New128()
	return vhll, nil
}

/*
Reset ...
*/
func (vhll *VirtualHyperLogLog) Reset() {
	vhll.registers.reset()
}

/*
GetRegisterValueDistribution ...
*/
func (vhll *VirtualHyperLogLog) GetRegisterValueDistribution(dist []float64) error {
	if dist == nil || len(dist) < 32 {
		return errors.New("distribution can't be < 32")
	}
	dist = make([]float64, len(dist), len(dist))
	for i := uint(0); i < vhll.physicalM; i++ {
		val := vhll.registers.get(i)
		dist[val]++
	}
	for i, val := range dist {
		dist[i] = val / float64(vhll.virtualM)
	}
	return nil
}
