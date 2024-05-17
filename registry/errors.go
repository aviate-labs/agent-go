package registry

import "fmt"

type ClientDataProviderQueryFailedError struct {
	Source DataProviderError
}

func (e ClientDataProviderQueryFailedError) Error() string {
	return fmt.Sprintf("failed to query data provider: %s", e.Source)
}

func (ClientDataProviderQueryFailedError) registryClientError() {}

type ClientDecodeError struct {
	Err string
}

func (e ClientDecodeError) Error() string {
	return fmt.Sprintf("failed to decode registry contents: %s", e.Err)
}

func (ClientDecodeError) registryClientError() {}

type ClientError interface {
	registryClientError()
	error
}

type ClientPollLockFailedError struct {
	Err string
}

func (e ClientPollLockFailedError) Error() string {
	return fmt.Sprintf("failed to acquire poll lock: %s", e.Err)
}

func (ClientPollLockFailedError) registryClientError() {}

type ClientPollingLatestVersionFailedError struct {
	Retries uint
}

func (e ClientPollingLatestVersionFailedError) Error() string {
	return fmt.Sprintf("failed to report the same version twice after %d times", e.Retries)
}

func (ClientPollingLatestVersionFailedError) registryClientError() {}

type ClientVersionNotAvailableError struct {
	Version Version
}

func (e ClientVersionNotAvailableError) Error() string {
	return fmt.Sprintf("the requested version is not available locally: %d", e.Version)
}

func (ClientVersionNotAvailableError) registryClientError() {}

type DataProviderError interface {
	dataProviderError()
	error
}

// DataProviderTimeoutError occurs when the registry transport client times out.
type DataProviderTimeoutError struct{}

func (DataProviderTimeoutError) Error() string {
	return "registry transport client timed out"
}

func (DataProviderTimeoutError) dataProviderError() {}

// DataProviderTransferError occurs when using registry transfer.
type DataProviderTransferError struct {
	Source string
}

func (e DataProviderTransferError) Error() string {
	return fmt.Sprintf("registry transport client failed to fetch registry update from registry canister: %s", e.Source)
}

func (DataProviderTransferError) dataProviderError() {}
