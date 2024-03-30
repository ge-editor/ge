//go:build ja_JP

// Implements te/file/Encorder interface

package charsetencoder

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/saintfish/chardet"
)

// Custom error
// Decoded messages
var (
	err_mac       = errors.New("Normalized from UTF-8-mac to UTF-8")
	err_shift_jis = errors.New("Encoded from ShiftJIS to UTF-8")
	err_euc_jp    = errors.New("Encoded from EUC-JP to UTF-8")

	err_unknown_encoding = errors.New("Unknown encoding")

	utf8     = "UTF-8"
	shiftjis = "Shift_JIS"
	eucjp    = "EUC-JP"
)

type Charencorder struct{}

func (e *Charencorder) Encoder(encoding string, b *[]byte) error {
	var err error
	switch encoding {
	case "UTF-8":
		return nil
	case "Shift_JIS":
		// Shift-JIS から UTF-8 に変換
		b, err = japanese.ShiftJIS.NewDecoder().Decode(bytes.NewReader(b))
		return err
	case "EUC-JP":
		// UTF-8 から EUC-JP に変換
		b, err = japanese.EUCJP.NewEncoder().Encode(utf8)
		return err
	}
	return err_unknown_encoding
}

// Read the specified number of bytes from the path file and Determine character code
func (e *Charencorder) GuessCharset(path string, bytes int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b := make([]byte, bytes)
	_, err = f.Read(b)
	if err != nil {
		return "", err
	}

	res, err := chardet.NewTextDetector().DetectBest(*b)
	return res.Charset, err
}

// Return Encoding string and error
func (e *Charencorder) Decoder(b *[]byte) (string, error) {
	res, err := chardet.NewTextDetector().DetectBest(*b)
	if err != nil {
		return "", err
	}

	switch res.Charset {
	case "UTF-8":
		// NFC (Normalization Form Composition), from UTF-8-mac to UTF-8
		nfc := norm.NFC.Bytes(*b)
		if bytes.Equal(*b, nfc) {
			nfc = nil
			return res.Charset, nil
		}
		*b = nfc
		nfc = nil
		return err_mac
	case "Shift_JIS":
		if *b, err = ioutil.ReadAll(transform.NewReader(bytes.NewReader(*b),
			japanese.ShiftJIS.NewDecoder())); err == nil {
			return res.Charset, err_shift_jis
		}
	case "EUC-JP":
		if *b, err = ioutil.ReadAll(transform.NewReader(bytes.NewReader(*b),
			japanese.EUCJP.NewDecoder())); err == nil {
			return res.Charset, err_euc_jp
		}
	}
	return err
}

func (e *Charencorder) IsDecodedMessage(err error) bool {
	return err == err_mac || err == err_shift_jis || err == err_euc_jp
}
