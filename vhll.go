package vhll

import (
	"errors"
	"math"

	metro "github.com/dgryski/go-metro"
)

func alpha(m float64) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	}
	return 0.7213 / (1 + 1.079/m)
}

func zeros(registers []uint8) (z float64) {
	for _, val := range registers {
		if val == 0 {
			z++
		}
	}
	return z
}

func beta(ez float64) float64 {
	zl := math.Log(ez + 1)
	return -0.370393911*ez +
		0.070471823*zl +
		0.17393686*math.Pow(zl, 2) +
		0.16339839*math.Pow(zl, 3) +
		-0.09237745*math.Pow(zl, 4) +
		0.03738027*math.Pow(zl, 5) +
		-0.005384159*math.Pow(zl, 6) +
		0.00042419*math.Pow(zl, 7)
}

// Calculate the position of the leftmost 1-bit.
func rho(val uint64) (r uint8) {
	for val&0x8000000000000000 == 0 {
		val <<= 1
		r++
	}
	return r + 1
}

func hash(e []byte) uint64 {
	return metro.Hash64(e, 1337)
}

func sumAndZeros(register []uint8) (float64, float64) {
	ez := 0.0
	sum := 0.0
	for _, val := range register {
		sum += 1.0 / math.Pow(2.0, float64(val))
		if val == 0 {
			ez++
		}
	}
	return sum, ez
}

// VHLL ...
type VHLL struct {
	M      []uint8
	m      uint64
	s      uint64
	log2s  uint64
	mAlpha float64
	sAlpha float64
}

func (v *VHLL) hashi(i uint64, f []byte) uint64 {
	return metro.Hash64(f, i) % v.m
}

// NewVHLL ...
func NewVHLL(precision, vPrecision uint8) (*VHLL, error) {
	if precision < 9 {
		return nil, errors.New("precision needs to be >= 9")
	}
	if vPrecision < 8 || vPrecision > 12 {
		return nil, errors.New("virtual precision needs to be >= 8 and <= 10")
	}
	if precision < vPrecision {
		return nil, errors.New("virtual precision needs to be > precision")
	}
	m := uint64(math.Pow(2, float64(precision)))
	s := uint64(math.Pow(2, float64(vPrecision)))
	return &VHLL{
		M:      make([]uint8, m, m),
		m:      m,
		s:      s,
		log2s:  uint64(vPrecision),
		mAlpha: alpha(float64(m)),
		sAlpha: alpha(float64(s)),
	}, nil
}

// Insert ...
func (v *VHLL) Insert(f []byte, e []byte) {
	he := hash(e)
	p := he % v.s
	q := he << v.log2s
	r := rho(q)
	index := metro.Hash64(f, p) % v.m
	if r > v.M[index] {
		v.M[index] = r
	}
}

// Estimate ...
func (v *VHLL) Estimate(f []byte) uint64 {
	M := make([]uint8, v.s, v.s)
	for i := range M {
		index := metro.Hash64(f, uint64(i)) % v.m
		M[i] = v.M[index]
	}

	sum, ez := sumAndZeros(M)
	s := float64(v.s)
	beta := beta(ez)
	ns := (v.sAlpha * s * (s - ez) / (beta + sum))

	// estimate error
	m := float64(v.m)
	n := float64(v.totalCardinality())
	e := ns - (s * n / m)

	// rounding
	return uint64(e + 0.5)
}

func (v *VHLL) totalCardinality() uint64 {
	sum, ez := sumAndZeros(v.M)
	m := float64(len(v.M))
	beta := beta(ez)
	return uint64(v.mAlpha * m * (m - ez) / (beta + sum))
}
