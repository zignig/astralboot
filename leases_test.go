package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLease(t *testing.T) {
	fmt.Println("---- LOCAL TESTS ------")
	l := &LeaseList{}
	l.Leases = append(l.Leases, &Lease{})
	fmt.Println(l)
	e, err := json.MarshalIndent(&l, "", "  ")
	fmt.Println(string(e), err)
}
