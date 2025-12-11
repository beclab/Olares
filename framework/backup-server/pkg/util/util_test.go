package util

import (
	"fmt"
	"testing"
	"time"
)

func TestS(t *testing.T) {
	// 1970-01-01 20:03:00
	// 43380000
	var a = fmt.Sprintf("1970-01-01 %s:00", "15:56")
	var b, _ = time.ParseInLocation(DateFormat, a, time.Local)
	fmt.Println("---1---", b.UnixMilli())

}

func TestX(t *testing.T) {
	c, _ := ParseTimestampToLocal("1744898760")
	fmt.Println("---c---", c)
}

func TestMD5(t *testing.T) {
	var a = MD5("/Files/Home/Documents")
	fmt.Println(a)
	var b = Base64encode([]byte(a))
	fmt.Println(b)

}
