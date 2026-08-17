package idl

import "fmt"

// DefaultDecodingQuota bounds the total work one decode may perform. The cost
// model follows the reference implementation: every value costs, so a bomb is
// caught whether or not its elements occupy bytes on the wire.
//
// A legitimate payload sits orders of magnitude below this; the ceiling only has
// to be low enough that a hostile one cannot exhaust memory first.
const DefaultDecodingQuota = 1 << 22

// Per-value costs, mirroring the reference implementation's cost model. The
// constant per-element charge is what bounds a vector whose elements occupy no
// bytes at all: its length alone spends the quota.
const (
	costValue     = 1 // any scalar
	costComposite = 2 // opt, record, variant, vec overhead
	costVecElem   = 3 // per element, before the element's own cost
)

// Budget is the total work allowance for one decode. It is shared by every type
// in that decode rather than held per composite, so neither nesting nor an
// intervening opt or record can widen it: every value drawn anywhere spends from
// the same pool.
//
// A nil *Budget is unmetered, which is what makes it safe to pass along paths
// that have no quota to enforce (encoding, Read).
type Budget struct {
	remaining int
	unlimited bool
}

func NewBudget(quota int) *Budget {
	return &Budget{remaining: quota}
}

// NewUnlimitedBudget returns a budget that never runs out, for payloads known to
// be well-formed but larger than a quota allows. Prefer it over a nil Budget at
// a call site: nil reads as "unspecified", which is the opposite of what it
// means. Decoding an untrusted payload this way can exhaust memory.
func NewUnlimitedBudget() *Budget {
	return &Budget{unlimited: true}
}

// Remaining reports the work left, and whether the budget is bounded at all.
func (b *Budget) Remaining() (int, bool) {
	if b == nil || b.unlimited {
		return 0, false
	}
	return b.remaining, true
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
