package utils

import (
	"bytes"
	"io/ioutil"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(data []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	return ioutil.ReadAll(reader)
}

func Utf16ToUtf8(data []byte) string {
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
	}

	runes := utf16.Decode(u16s)
	buf := make([]byte, 0, len(runes)*3)
	for _, r := range runes {
		buf = append(buf, make([]byte, utf8.RuneLen(r))...)
		utf8.EncodeRune(buf[len(buf)-utf8.RuneLen(r):], r)
	}
	return string(buf)
}
