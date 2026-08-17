package idl

import "fmt"

// DefaultDecodingQuota bounds the total work one decode may perform. The cost
// model follows the reference implementation: every value costs, so a bomb is
// caught whether or not its elements occupy bytes on the wire.
//
// IC message size is implementation-defined, in practice 2 MiB of payload
// against a 3 MiB message ceiling. The densest legal payload is a vec nat8, one
// wire byte per element at costVecElem+costValue, so this has to clear 4x the
// message ceiling to avoid rejecting traffic that used to decode. 1<<24 covers
// 4 MiB of that worst case while still refusing a bomb long before it can
// exhaust memory.
const DefaultDecodingQuota = 1 << 24

// Per-value costs, mirroring the reference implementation's cost model.
const (
	costValue     = 1 // any scalar
	costComposite = 2 // opt, record, variant, vec overhead
	costVecElem   = 3 // per element, before the element's own cost
)

// maxZeroWidthValues caps values that occupy no wire bytes at all. These need
// their own ceiling rather than the work quota: a quota has to scale with the
// payload to admit a large legal message, but a zero-width count has no
// relationship to payload size, so a few bytes could always claim the whole
// allowance. The pool is shared across one decode, so nesting cannot multiply
// past it.
const maxZeroWidthValues = 1 << 16

// Budget is the total work allowance for one decode. It is shared by every type
// in that decode rather than held per composite, so neither nesting nor an
// intervening opt or record can widen it: every value drawn anywhere spends from
// the same pool.
//
// A nil *Budget is unmetered, which is what makes it safe to pass along paths
// that have no quota to enforce (encoding, Read).
type Budget struct {
	remaining int
	zeroWidth int
	unlimited bool
}

func NewBudget(quota int) *Budget {
	return &Budget{remaining: quota, zeroWidth: maxZeroWidthValues}
}

// SpendZeroWidth charges n values that carry no wire bytes, against a ceiling
// separate from the work quota.
func (b *Budget) SpendZeroWidth(n int) error {
	if b == nil || b.unlimited {
		return nil
	}
	if n > b.zeroWidth {
		return fmt.Errorf("zero-width values exceed limit: need %d, %d remaining", n, b.zeroWidth)
	}
	b.zeroWidth -= n
	return b.Spend(n * costVecElem)
}

// NewUnlimitedBudget returns a budget that never runs out, for payloads known to
// be well-formed but larger than a quota allows. Prefer it over a nil Budget at
// a call site: nil reads as "unspecified", which is the opposite of what it
// means. Decoding an untrusted payload this way can exhaust memory.
func NewUnlimitedBudget() *Budget {
	return &Budget{unlimited: true}
}

// RemainingZeroWidth reports the zero-width values left, and whether they are
// bounded at all.
func (b *Budget) RemainingZeroWidth() (int, bool) {
	if b == nil || b.unlimited {
		return 0, false
	}
	return b.zeroWidth, true
}

// Spend charges n against the budget, reporting an error once a payload has
// asked for more work than any payload of its size could justify.
func (b *Budget) Spend(n int) error {
	if b == nil || b.unlimited {
		return nil
	}
	if n > b.remaining {
		return fmt.Errorf("decoding quota exhausted: need %d, %d remaining", n, b.remaining)
	}
	b.remaining -= n
	return nil
}
