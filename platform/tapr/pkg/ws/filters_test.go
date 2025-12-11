package ws

import (
	"fmt"
	"testing"
)

var server = &Server{
	users: map[string]*User{
		"alice": {
			name: "alice",
			conns: map[string]*Client{
				"c-1": {},
				"c-2": {},
				"c-3": {},
				"c-4": {},
			},
		},
		"bob": {
			name: "bob",
			conns: map[string]*Client{
				"c-10": {},
				"c-11": {},
				"c-12": {},
				"c-13": {},
			},
		},
		"cash": {
			name: "cash",
			conns: map[string]*Client{
				"c-15": {},
				"c-16": {},
			},
		},
	},
}

func TestFilter1(t *testing.T) {
	var f = NewFilter(server)
	var r = f.FilterByUsers([]string{"alice", "cash"}).FilterByConnIds([]string{}).Result() // []
	fmt.Println("---conns-1---", r)

	f = NewFilter(server)
	r = f.FilterByConnIds([]string{"c-3", "c-11"}).FilterByUsers([]string{"alice", "bob"}).Result() // [c-3 c-11]
	fmt.Println("---conns-2---", r)

	f = NewFilter(server)
	r = f.FilterByUsers([]string{"alice"}).FilterByConnIds([]string{"c-1"}).Result() // [c-1]
	fmt.Println("---conns-3---", r)

	f = NewFilter(server)
	r = f.FilterByConnIds([]string{"c-1", "c-16"}).FilterByUsers([]string{"cash"}).Result() // [c-16]
	fmt.Println("---conns-4---", r)

	f = NewFilter(server)
	r = f.FilterByUsers([]string{"alice"}).FilterByConnIds([]string{"c-10"}).Result() // []
	fmt.Println("---conns-5---", r)

	f = NewFilter(server)
	r = f.FilterByUsers([]string{""}).FilterByConnIds([]string{"c-1", "c-2", "c-3", "c-4", "c-10", "c-11", "c-12", "c-13", "c-15"}).Result() // []
	fmt.Println("---conns-6---", r)

	f = NewFilter(server)
	r = f.FilterByConnIds([]string{"c-1", "c-2", "c-3", "c-4", "c-10", "c-11", "c-12", "c-13", "c-15"}).FilterByUsers([]string{"alice"}).Result()
	// [c-2 c-3 c-4 c-1]
	fmt.Println("---conns-7---", r)

	f = NewFilter(server)
	r = f.FilterByUsers([]string{"bob", "cash"}).FilterByConnIds([]string{""}).Result() // []
	fmt.Println("---conns-8---", r)
}
