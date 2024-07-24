package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Custom Types

/*
 * Alias Types
 */
// No pueden asociarse metodos

type (
	ID        = uint64
	UUIDAlias = [32]byte
	UserIDs   = []ID
	bytes     = []byte
)

func aliasTypes() {
	var myID ID
	myID = 1

	var myIDUint64 = myID
	fmt.Println(myIDUint64)
}

// Invalid syntax
//func (id ID) MyMethod() {
//
//}

/*
 * Wrapper Types
 */
// Si pueden asociarse metodos

type IDWrapper uint64

func (id IDWrapper) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

type UUID [32]byte

func (id UUID) String() string {
	sb := strings.Builder{}

	for _, b := range id {
		sb.WriteByte(b)
	}

	return sb.String()
}

func uuidExample() {
	json.Unmarshal([]uint8{1, 2, 3}, nil)
	json.Unmarshal(bytes{1, 2, 3}, nil)
}

type (
	Interface interface {
		Method()
	}

	Structure struct {
		ID    uint64 `tag1:"tagValue1|tagValue2|tagValue3" tag2:"tagValue" tag3:"tagValue" gorm:"<-create;<-update;<-delete"`
		Value string `binding:"email"` // go-validator
	}

	Wrapper uint64

	WrapperFunc func()
)

func main() {

}
