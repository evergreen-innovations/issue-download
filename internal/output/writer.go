package output

import "io"

// Writer holds onto the first error that occurs.
// This means we don't have to check the error
// after each write. Just the last one
type writer struct {
	w   io.Writer
	err error
}

func (w *writer) WriteString(str string) {
	// If we already have an error don't try to write
	if w.err != nil {
		return
	}

	_, err := w.w.Write([]byte(str))
	// Save the error
	if err != nil && w.err == nil {
		w.err = err
	}
}

func (w *writer) Error() error {
	return w.err
}
