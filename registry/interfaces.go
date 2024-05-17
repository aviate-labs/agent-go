package registry

import "time"

func GetValue(c RegistryClient, key string, version Version) (*[]byte, error) {
	vv, err := c.GetVersionedValue(key, version)
	if err != nil {
		return nil, err
	}
	return vv.Value, nil
}

type RegistryClient interface {
	GetVersionedValue(key string, version Version) (VersionedRecord[[]byte], error)
	GetKeyFamily(keyPrefix string, version Version) ([]string, error)
	GetLatestVersion() (Version, error)
	GetVersionTimestamp(version Version) (*int64, error)
}

type RegistryDataProvider interface {
	GetUpdatesSince(version Version) ([]TransportRecord, error)
}

type TransportRecord = VersionedRecord[[]byte]

func EmptyZeroRecord(key string) TransportRecord {
	return TransportRecord{
		Key:     key,
		Version: ZeroRegistryVersion,
		Value:   nil,
	}
}

type Version uint64

// Reference: https://github.com/dfinity/ic/blob/master/rs/interfaces/registry/src/lib.rs

const (
	// ZeroRegistryVersion is the version number of the empty registry.
	ZeroRegistryVersion Version = 0
	// PollingPeriod is the period at which the local store is polled for updates.
	PollingPeriod = 5 * time.Second
)

// VersionedRecord is a key-value pair with a version.
type VersionedRecord[T any] struct {
	// Key of the record.
	Key string
	// Version at which this record was created.
	Version Version
	// Value of the record. If the record was deleted in this version, this field is nil.
	Value *T
}
