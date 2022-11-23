package candid

const pow32 = 4294967296

// HashId calculates the the hash of the given id.
//
// The hash function is specified as:
// hash(id) = ( Sum_(i=0..k) utf8(id)[i] * 223^(k-i) ) mod 2^32 where k = |utf8(id)|-1
func HashId(id string) (hash int) {
	for _, i := range []byte(id) {
		hash = (hash*223 + int(i)) % pow32
	}
	return
}
