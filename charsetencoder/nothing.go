//!go:build ja_JP

// Implements te/file/Encoder interface
// I haven't decided how to call te from a package.

package charsetencoder

type Charencorder struct{}

func (e *Charencorder) Encoder(encoding string, b *[]byte) error {
	return nil
}

// Read the specified number of bytes from the path file and Determine character code
func (e *Charencorder) GuessCharset(path string, bytes int) (string, error) {
	return "UTF-8", nil
}

func (e *Charencorder) Decoder(b *[]byte) (string, error) {
	return "UTF-8", nil
}

func (e *Charencorder) IsDecodedMessage(err error) bool {
	return false
}
