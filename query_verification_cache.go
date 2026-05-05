package agent

import (
	"context"
	"crypto/ed25519"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/niccolofant/agent-go/certification"
	"github.com/niccolofant/agent-go/certification/hashtree"
	"github.com/niccolofant/agent-go/principal"
)

type queryVerificationKeyCache struct {
	mu           sync.RWMutex
	subnets      map[string]*queryVerificationKeySet
	canisters    map[string]string
	ranges       []queryVerificationSubnetRange
	refreshing   map[string]bool
	refreshAfter time.Duration
	maxAge       time.Duration
}

type queryVerificationSubnetRange struct {
	subnetID principal.Principal
	ranges   certification.CanisterRanges
}

type queryVerificationKeySet struct {
	subnetID       principal.Principal
	canisterRanges certification.CanisterRanges
	keys           map[string]ed25519.PublicKey
	refreshAfter   time.Time
	validUntil     time.Time
}

func newQueryVerificationKeyCache(ingressExpiry time.Duration) *queryVerificationKeyCache {
	maxAge := 30 * time.Second
	if 0 < ingressExpiry && ingressExpiry < maxAge {
		maxAge = ingressExpiry / 2
		if maxAge <= 0 {
			maxAge = ingressExpiry
		}
	}

	refreshAfter := 5 * time.Second
	if maxAge < refreshAfter {
		refreshAfter = maxAge / 2
		if refreshAfter <= 0 {
			refreshAfter = maxAge
		}
	}

	return &queryVerificationKeyCache{
		subnets:      make(map[string]*queryVerificationKeySet),
		canisters:    make(map[string]string),
		refreshing:   make(map[string]bool),
		refreshAfter: refreshAfter,
		maxAge:       maxAge,
	}
}

// WarmQueryVerificationCache warms the signed query verification cache for the given canisters.
func (a Agent) WarmQueryVerificationCache(canisterIDs ...principal.Principal) error {
	return a.WarmQueryVerificationCacheContext(a.ctx, canisterIDs...)
}

// WarmQueryVerificationCacheContext warms the signed query verification cache for the given canisters.
func (a Agent) WarmQueryVerificationCacheContext(ctx context.Context, canisterIDs ...principal.Principal) error {
	if ctx == nil {
		ctx = a.ctx
	}
	for _, canisterID := range canisterIDs {
		keys, err := a.fetchQueryVerificationKeys(ctx, canisterID, nil)
		if err != nil {
			return err
		}
		a.storeQueryVerificationKeys(canisterID, keys)
	}
	return nil
}

func (a Agent) queryVerificationKeys(ctx context.Context, ecID principal.Principal, signatures []ResponseSignature) (*queryVerificationKeySet, error) {
	identities := signatureIdentities(signatures)
	if a.queryVerificationCache == nil {
		return a.fetchQueryVerificationKeys(ctx, ecID, identities)
	}

	if keys, stale := a.queryVerificationCache.get(ecID, identities); keys != nil {
		if stale {
			a.refreshQueryVerificationKeys(ecID, keys.identities())
		}
		return keys, nil
	}

	keys, err := a.fetchQueryVerificationKeys(ctx, ecID, identities)
	if err != nil {
		return nil, err
	}
	a.storeQueryVerificationKeys(ecID, keys)
	return keys, nil
}

func (a Agent) storeQueryVerificationKeys(ecID principal.Principal, keys *queryVerificationKeySet) {
	if a.queryVerificationCache != nil {
		a.queryVerificationCache.store(ecID, keys)
	}
}

func (a Agent) refreshQueryVerificationKeys(ecID principal.Principal, identities []principal.Principal) {
	if a.queryVerificationCache == nil {
		return
	}
	refreshKey, ok := a.queryVerificationCache.beginRefresh(ecID)
	if !ok {
		return
	}
	go func() {
		defer a.queryVerificationCache.endRefresh(refreshKey)
		keys, err := a.fetchQueryVerificationKeys(a.ctx, ecID, identities)
		if err != nil {
			a.logger.Printf("[AGENT] QUERY VERIFICATION CACHE refresh failed for %s: %v", ecID, err)
			return
		}
		a.queryVerificationCache.store(ecID, keys)
	}()
}

func (a Agent) fetchQueryVerificationKeys(ctx context.Context, ecID principal.Principal, identities []principal.Principal) (*queryVerificationKeySet, error) {
	certificate, err := a.readStateCertificateContext(ctx, ecID, [][]hashtree.Label{{hashtree.Label("subnet")}})
	if err != nil {
		return nil, err
	}

	subnetID := principal.MustDecode(certification.RootSubnetID)
	if certificate.Delegation != nil {
		subnetID = certificate.Delegation.SubnetId
	}

	keys, err := queryVerificationPublicKeys(certificate, subnetID)
	if err != nil {
		if len(identities) == 0 {
			return nil, err
		}
		keys, err = queryVerificationPublicKeysForIdentities(certificate, subnetID, identities)
		if err != nil {
			return nil, err
		}
	} else if !queryVerificationHasAll(keys, identities) {
		missingKeys, err := queryVerificationPublicKeysForIdentities(certificate, subnetID, identities)
		if err != nil {
			return nil, err
		}
		for k, v := range missingKeys {
			keys[k] = v
		}
	}

	canisterRanges := queryVerificationCanisterRanges(certificate, subnetID)
	now := time.Now()
	return &queryVerificationKeySet{
		subnetID:       subnetID,
		canisterRanges: canisterRanges,
		keys:           keys,
		refreshAfter:   now.Add(a.queryVerificationRefreshAfter()),
		validUntil:     now.Add(a.queryVerificationMaxAge()),
	}, nil
}

func queryVerificationPublicKeys(certificate *certification.Certificate, subnetID principal.Principal) (map[string]ed25519.PublicKey, error) {
	nodes, err := certificate.Tree.LookupSubTree(hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("node"))
	if err != nil {
		return nil, err
	}
	children, err := hashtree.AllChildren(nodes)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]ed25519.PublicKey, len(children))
	for _, node := range children {
		if len(node.Path) == 0 {
			continue
		}
		pk, err := hashtree.Lookup(node.Value, hashtree.Label("public_key"))
		if err != nil {
			return nil, err
		}
		publicKey, err := certification.PublicED25519KeyFromDER(pk)
		if err != nil {
			return nil, err
		}
		keys[principalCacheKey(principal.Principal{Raw: node.Path[0]})] = append(ed25519.PublicKey(nil), (*publicKey)...)
	}
	return keys, nil
}

func queryVerificationPublicKeysForIdentities(certificate *certification.Certificate, subnetID principal.Principal, identities []principal.Principal) (map[string]ed25519.PublicKey, error) {
	keys := make(map[string]ed25519.PublicKey, len(identities))
	for _, identity := range identities {
		pk, err := certificate.Tree.Lookup(
			hashtree.Label("subnet"),
			subnetID.Raw,
			hashtree.Label("node"),
			identity.Raw,
			hashtree.Label("public_key"),
		)
		if err != nil {
			return nil, err
		}
		publicKey, err := certification.PublicED25519KeyFromDER(pk)
		if err != nil {
			return nil, err
		}
		keys[principalCacheKey(identity)] = append(ed25519.PublicKey(nil), (*publicKey)...)
	}
	return keys, nil
}

func queryVerificationCanisterRanges(certificate *certification.Certificate, subnetID principal.Principal) certification.CanisterRanges {
	var canisterRanges certification.CanisterRanges
	rawCanisterRanges, err := certificate.Tree.Lookup(hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("canister_ranges"))
	if err == nil {
		_ = cbor.Unmarshal(rawCanisterRanges, &canisterRanges)
		return canisterRanges
	}
	if certificate.Delegation == nil {
		return nil
	}
	rawCanisterRanges, err = certificate.Delegation.Certificate.Tree.Lookup(hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("canister_ranges"))
	if err != nil {
		return nil
	}
	_ = cbor.Unmarshal(rawCanisterRanges, &canisterRanges)
	return canisterRanges
}

func queryVerificationHasAll(keys map[string]ed25519.PublicKey, identities []principal.Principal) bool {
	for _, identity := range identities {
		if _, ok := keys[principalCacheKey(identity)]; !ok {
			return false
		}
	}
	return true
}

func (a Agent) queryVerificationRefreshAfter() time.Duration {
	if a.queryVerificationCache == nil {
		return 5 * time.Second
	}
	return a.queryVerificationCache.refreshAfter
}

func (a Agent) queryVerificationMaxAge() time.Duration {
	if a.queryVerificationCache == nil {
		return 30 * time.Second
	}
	return a.queryVerificationCache.maxAge
}

func (c *queryVerificationKeyCache) get(ecID principal.Principal, identities []principal.Principal) (*queryVerificationKeySet, bool) {
	c.mu.RLock()
	subnetKey := c.subnetKeyForCanisterLocked(ecID)
	entry := c.subnets[subnetKey]
	now := time.Now()
	if entry == nil || !now.Before(entry.validUntil) || !entry.hasAll(identities) {
		c.mu.RUnlock()
		return nil, false
	}
	stale := !now.Before(entry.refreshAfter)
	entry = entry.clone()
	c.mu.RUnlock()
	return entry, stale
}

func (c *queryVerificationKeyCache) store(ecID principal.Principal, keys *queryVerificationKeySet) {
	canisterKey := principalCacheKey(ecID)
	subnetKey := principalCacheKey(keys.subnetID)
	c.mu.Lock()
	if existing := c.subnets[subnetKey]; existing != nil {
		merged := existing.clone()
		merged.refreshAfter = keys.refreshAfter
		merged.validUntil = keys.validUntil
		if len(keys.canisterRanges) != 0 {
			merged.canisterRanges = cloneCanisterRanges(keys.canisterRanges)
		}
		for k, v := range keys.keys {
			merged.keys[k] = append(ed25519.PublicKey(nil), v...)
		}
		keys = merged
	}
	c.subnets[subnetKey] = keys.clone()
	c.canisters[canisterKey] = subnetKey
	c.storeRangesLocked(keys.subnetID, keys.canisterRanges)
	c.mu.Unlock()
}

func (c *queryVerificationKeyCache) beginRefresh(ecID principal.Principal) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cacheKey := c.subnetKeyForCanisterLocked(ecID)
	if cacheKey == "" {
		cacheKey = principalCacheKey(ecID)
	}
	if c.refreshing[cacheKey] {
		return "", false
	}
	c.refreshing[cacheKey] = true
	return cacheKey, true
}

func (c *queryVerificationKeyCache) endRefresh(cacheKey string) {
	c.mu.Lock()
	delete(c.refreshing, cacheKey)
	c.mu.Unlock()
}

func (c *queryVerificationKeyCache) subnetKeyForCanisterLocked(ecID principal.Principal) string {
	canisterKey := principalCacheKey(ecID)
	if subnetKey := c.canisters[canisterKey]; subnetKey != "" {
		return subnetKey
	}
	for _, r := range c.ranges {
		if r.ranges.InRange(ecID) {
			return principalCacheKey(r.subnetID)
		}
	}
	return ""
}

func (c *queryVerificationKeyCache) storeRangesLocked(subnetID principal.Principal, ranges certification.CanisterRanges) {
	if len(ranges) == 0 {
		return
	}
	subnetKey := principalCacheKey(subnetID)
	for i, r := range c.ranges {
		if principalCacheKey(r.subnetID) == subnetKey {
			c.ranges[i].ranges = cloneCanisterRanges(ranges)
			return
		}
	}
	c.ranges = append(c.ranges, queryVerificationSubnetRange{
		subnetID: principal.Principal{Raw: append([]byte(nil), subnetID.Raw...)},
		ranges:   cloneCanisterRanges(ranges),
	})
}

func (s *queryVerificationKeySet) publicKey(identity principal.Principal) (ed25519.PublicKey, bool) {
	key, ok := s.keys[principalCacheKey(identity)]
	if !ok {
		return nil, false
	}
	return key, true
}

func (s *queryVerificationKeySet) hasAll(identities []principal.Principal) bool {
	for _, identity := range identities {
		if _, ok := s.keys[principalCacheKey(identity)]; !ok {
			return false
		}
	}
	return true
}

func (s *queryVerificationKeySet) identities() []principal.Principal {
	identities := make([]principal.Principal, 0, len(s.keys))
	for raw := range s.keys {
		identities = append(identities, principal.Principal{Raw: []byte(raw)})
	}
	return identities
}

func (s *queryVerificationKeySet) clone() *queryVerificationKeySet {
	keys := make(map[string]ed25519.PublicKey, len(s.keys))
	for k, v := range s.keys {
		keys[k] = append(ed25519.PublicKey(nil), v...)
	}
	return &queryVerificationKeySet{
		subnetID:       principal.Principal{Raw: append([]byte(nil), s.subnetID.Raw...)},
		canisterRanges: cloneCanisterRanges(s.canisterRanges),
		keys:           keys,
		refreshAfter:   s.refreshAfter,
		validUntil:     s.validUntil,
	}
}

func cloneCanisterRanges(ranges certification.CanisterRanges) certification.CanisterRanges {
	if len(ranges) == 0 {
		return nil
	}
	clone := make(certification.CanisterRanges, len(ranges))
	for i, r := range ranges {
		clone[i] = certification.CanisterRange{
			From: principal.Principal{Raw: append([]byte(nil), r.From.Raw...)},
			To:   principal.Principal{Raw: append([]byte(nil), r.To.Raw...)},
		}
	}
	return clone
}

func signatureIdentities(signatures []ResponseSignature) []principal.Principal {
	seen := make(map[string]struct{}, len(signatures))
	identities := make([]principal.Principal, 0, len(signatures))
	for _, signature := range signatures {
		key := principalCacheKey(signature.Identity)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		identities = append(identities, signature.Identity)
	}
	return identities
}

func principalCacheKey(p principal.Principal) string {
	return string(p.Raw)
}
