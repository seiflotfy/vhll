package vhll

type registerSet struct {
	Count uint
	Size  uint
	m     []uint8
}

func newRegisterSet(count uint) *registerSet {
	r := &registerSet{Count: count, Size: count / 8}

	r.reset()
	return r
}

func newRegisterSetWithValues(count uint, initialValues []uint8) *registerSet {
	rs := &registerSet{Size: count / 8, Count: count}
	rs.m = initialValues
	rs.Size = uint(len(rs.m))
	return rs
}

func (rs *registerSet) reset() {
	rs.m = make([]uint8, rs.Size, rs.Size)
}

func (rs *registerSet) set(pos uint, val uint8) {
	rs.m[pos/8] = val
}

func (rs *registerSet) get(pos uint) uint8 {
	return rs.m[pos/8]
}

func (rs *registerSet) updateIfGreater(pos uint, val uint8) bool {
	if rs.m[pos/8] < val {
		rs.m[pos/8] = val
		return true
	}
	return false
}

func (rs *registerSet) merge(ors *registerSet) {
	for i, val := range ors.m {
		if val > rs.m[i] {
			rs.m[i] = val
		}
	}
}
