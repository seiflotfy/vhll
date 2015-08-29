package vhll

type registerSet struct {
	Count uint
	Size  uint
	M     []uint8
}

func newRegisterSet(count uint) *registerSet {
	r := &registerSet{Count: count, Size: count}

	r.reset()
	return r
}

func (rs *registerSet) reset() {
	rs.M = make([]uint8, rs.Size, rs.Size)
}

func (rs *registerSet) set(pos uint, val uint8) {
	rs.M[pos] = val
}

func (rs *registerSet) get(pos uint) uint8 {
	return rs.M[pos]
}

func (rs *registerSet) updateIfGreater(pos uint, val uint8) bool {
	if rs.M[pos] < val {
		rs.M[pos] = val
		return true
	}
	return false
}

func (rs *registerSet) merge(ors *registerSet) {
	for i, val := range ors.M {
		if val > rs.M[i] {
			rs.M[i] = val
		}
	}
}
