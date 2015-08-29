package vhll

import (
	"errors"
	"math"
	"strconv"

	"github.com/dgryski/go-spooky"
)

var (
	exp32 = math.Pow(2, 32)
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
	m := uint(math.Pow(2, float64(log2m)))

	var alphaMM float64

	// See the paper.
	switch log2m {
	case 4:
		alphaMM = 0.673 * float64(m*m)
		break
	case 5:
		alphaMM = 0.697 * float64(m*m)
		break
	case 6:
		alphaMM = 0.709 * float64(m*m)
	default:
		alphaMM = (0.7213 / (1 + 1.079/float64(m))) * float64(m*m)
	}

	return alphaMM
}

func round(f float64) uint {
	return uint(f + 0.5)
}

// Calculate the position of the leftmost 1-bit.
func getLeadingZeros(val uint64, max uint32) uint8 {
	r := uint32(1)
	for val&0x8000000000000000 == 0 && r <= max {
		r++
		val <<= 1
	}
	return uint8(r)
}

/*
VirtualHyperLogLog a Highly Compact Virtual Maximum Likelihood Sketche
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
	noiseCorrector  float64
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
NewForLog2m creates a new VirtualHyperLogLog with a given log2m, which needs to be dividable by 8
*/
func NewForLog2m(log2m uint) (*VirtualHyperLogLog, error) {
	return new(log2m, newRegisterSet(uint(math.Pow(2, float64(log2m)))))
}

/*
New ...
*/
func new(physicalLog2m uint, registers *registerSet) (*VirtualHyperLogLog, error) {
	vhll := &VirtualHyperLogLog{}
	vhll.registers = registers
	vhll.physicalLog2m = physicalLog2m
	vhll.physicalAlphaMM = getAlphaMM(physicalLog2m)
	vhll.physicalM = uint(math.Pow(2, float64(physicalLog2m)))

	if physicalLog2m < 7 {
		return nil, errors.New("physicalLog2m needs to be >= 7")
	}

	vhll.virtualLog2m = getVirtualEstimatorSize(physicalLog2m)
	vhll.virtualAlphaMM = getAlphaMM(vhll.virtualLog2m)

	vhll.virtualM = uint(math.Pow(2, float64(vhll.virtualLog2m)))
	vhll.virtualCa = mAlpha[vhll.virtualLog2m]
	vhll.noiseCorrector = 1
	return vhll, nil
}

/*
Reset clears all data from the struct
*/
func (vhll *VirtualHyperLogLog) Reset() {
	vhll.noiseCorrector = 1
	vhll.registers.reset()
}

func (vhll *VirtualHyperLogLog) getPhysicalRegisterFromVirtualRegister(id []byte, virtual uint) uint {
	idx := uint(spooky.Hash64(id))
	n := (idx+13)*104729 + virtual
	h1 := spooky.Hash64([]byte(strconv.Itoa(int(n))))
	return uint((uint(h1) & 0xFFFFFFFFFFFF) % vhll.physicalM)
}

/*
Add pushes data to the vritual hyperloglog of a flow 'id'
*/
func (vhll *VirtualHyperLogLog) Add(id []byte, data []byte) bool {
	data = append(data, id...)
	h1 := spooky.Hash64(data)
	virtualRegister := h1 >> (64 - vhll.virtualLog2m)
	r := getLeadingZeros(((h1<<vhll.virtualLog2m)|(1<<(vhll.virtualLog2m-1))+1)+1, 32)
	physicalRegister := vhll.getPhysicalRegisterFromVirtualRegister(id, uint(virtualRegister))
	return vhll.registers.updateIfGreater(physicalRegister, r)
}

/*
GetTotalCardinality returns cardinality across flows
*/
func (vhll *VirtualHyperLogLog) GetTotalCardinality() uint64 {
	registerSum := float64(0)
	count := vhll.registers.Count
	zeros := 0.0

	totalCardinalityFromPhySpace := 0
	for _, val := range vhll.registers.M {
		registerSum += 1.0 / float64(uint(1)<<val)
		if val == 0 {
			zeros++
		}
	}

	estimate := vhll.physicalAlphaMM * (1 / registerSum)
	if estimate <= (5.0/2.0)*float64(count) {
		totalCardinalityFromPhySpace = int(round(float64(count) * math.Log(float64(count)/zeros)))
	} else {
		totalCardinalityFromPhySpace = int(round(estimate))
	}

	//vhll.noiseCorrector = float64(vhll.totalCardinality) / float64(totalCardinalityFromPhySpace)
	return uint64(round(float64(totalCardinalityFromPhySpace)))
}

func (vhll *VirtualHyperLogLog) getNoiseMean() float64 {
	nhat := vhll.GetTotalCardinality()
	m := vhll.physicalM
	s := vhll.virtualM
	return float64(uint(nhat)) * float64(s/m)
}

/*
GetCardinality return the cardinality of a flow 'id'
*/
func (vhll *VirtualHyperLogLog) GetCardinality(id []byte) float64 {

	physicalCardinality := vhll.GetTotalCardinality()
	registerSum := float64(0)
	zeros := float64(0)
	for j := uint(0); j < vhll.virtualM; j++ {
		phyReg := vhll.getPhysicalRegisterFromVirtualRegister(id, j)
		val := vhll.registers.get(phyReg)
		registerSum += 1.0 / float64(uint(1)<<val)
		if val == 0 {
			zeros++
		}
	}
	estimate := float64(vhll.virtualAlphaMM) * (1 / registerSum)

	virtualCardinality := float64(round(estimate))

	vp := float64(1.0 * vhll.physicalM * vhll.virtualM / (vhll.physicalM - vhll.virtualM))
	result := float64(0)
	noiseMean := vhll.getNoiseMean()

	if vhll.virtualLog2m >= vhll.physicalLog2m-6 {
		result = float64(round(vp * ((virtualCardinality)/float64(vhll.virtualM) - float64(physicalCardinality/uint64(vhll.physicalM)))))
	} else {
		result = float64(round(vp * ((virtualCardinality)/float64(vhll.virtualM) - float64(physicalCardinality)/vhll.noiseCorrector/float64(vhll.physicalM))))
		if result-(1.2*noiseMean) > 0 {
			result = float64(round(vp * ((virtualCardinality)/float64(vhll.virtualM) - float64(physicalCardinality/uint64(vhll.physicalM)))))
		}
	}
	return result
}
