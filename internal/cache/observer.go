package cache

// Observer receives cache lifecycle events. All methods must be safe for concurrent use.
type Observer interface {
	OnHit()    // value returned from cache
	OnMiss()   // key not present
	OnExpire() // key found but TTL elapsed (item removed)
	OnEvict()  // oldest item removed to make room
	OnAdd()    // new key inserted
}
