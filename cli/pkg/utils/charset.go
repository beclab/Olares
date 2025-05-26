package utils

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	CharsetWindows1252 = "windows-1252"
	CharsetISO88591    = "ISO-8859-1"
	CharsetGB18030     = "GB-18030"
)

type GB18030 struct {
	data []byte
}

func (c *GB18030) Utf8() (string, error) {
	decoder := simplifiedchinese.GB18030.NewDecoder()
	unicodeData, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(c.data), decoder))
	if err != nil {
		return string(c.data), err
	}

	return string(unicodeData), nil
}

type Windows1252 struct {
	data []byte
}

func (c *Windows1252) Utf8() (string, error) {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	unicodeData, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(c.data), decoder))
	if err != nil {
		return string(c.data), err
	}

	return string(unicodeData), nil
}

func (c *Windows1252) Gb2312() (string, error) {
	decoder := charmap.Windows1252.NewDecoder()
	unicodeData, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(c.data), decoder))

	encoder := simplifiedchinese.GB18030.NewEncoder()
	gb2312Data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(unicodeData), encoder))
	if err != nil {
		return string(c.data), err
	}

	return string(gb2312Data), nil
}

type ISO8859 struct {
	data []byte
}

func (c *ISO8859) Utf8() (string, error) {
	decoder := charmap.ISO8859_1.NewDecoder()
	utf8Data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(c.data), decoder))
	if err != nil {
		return string(c.data), err
	}

	return string(utf8Data), nil
}

func CharsetConverts(charset string, data []byte, c Charset) (string, error) {
	switch {
	case c == DEFAULT:
		switch charset {
		case CharsetWindows1252, CharsetISO88591:
			return (&Windows1252{data: data}).Utf8()
		// case CharsetISO88591:
		// 	return (&ISO8859{data: data}).Utf8()
		case CharsetGB18030:
			return (&GB18030{data: data}).Utf8()
		default:
			return string(data), nil
		}
	default:
		return string(data), nil
	}
}
