// Package fault names persistence boundaries. Hooks may terminate a test process;
// production callers leave Hook nil. A returned error aborts the operation.
package fault

type Hook func(point string) error

func (h Hook) Hit(point string) error {
	if h != nil {
		return h(point)
	}
	return nil
}
