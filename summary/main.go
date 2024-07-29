package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	//Variables
	var variable01 int
	variable02 := uint32(2)
	var variable03 uint64 = 3
	// var x = variable03.(uint32)

	var variable04 any = ""
	var variable05, _ = variable04.(uint32)
	print(variable01, variable02, variable03, variable05)

	//Constants
	const constanteInt = 1
	const constanteUInt uint = 1

	//For
	// Forever -> While
	cont := 0
	for {
		println("Holi")
		cont++
		if cont == 5 {
			break
		}
	}

	// For i traditional
	for i := uint(0); i <= 5; i++ {
		println("Holi")
	}

	// For Range (index, value := range collection)
	for index, value := range []string{"holi", "chau"} {
		println(index, value)
	}

	// For Range Channels
	//for value := range make(chan uint64) {
	//	println(value)
	//}

	//If/Else
	if true {

	} else {

	}

	// Ternary
	// result = condition ? value1 : value2
	condition := true
	value := "value1"
	if condition {
		value = "value2"
	}
	println(value)

	//Switch
	// Switch - Case
	var switchVar any
	switch switchVar {
	case "1":
		println("Hola mundo")
		// break -> automatic
		fallthrough // continue validating
	case "2", "3":
		fallthrough
	case "5":
		return
	default:
		println("Hola mundo")
	}

	// TODO: revisar
	//var structValue struct{ ID uint64 }
	//switch structValue {
	//case struct{ ID uint64 }{
	//	ID: 1,
	//}:
	//	println("Hola mundo")
	//}

	// Switch - Type
	var myVariable any = "hola mundo"
	switch myValue := myVariable.(type) {
	case int, uint, float32, float64:
		println(myValue)
	case string:
		println(myValue)
	}

	//Arrays - Slices
	array01 := [4]uint{1, 2, 4, 4}
	array02 := [...]uint{1, 2, 4, 4, 5, 6}
	fmt.Println(array01, array02)

	// array02 = append(array02, 1) No compile

	slice01 := []int{1, 2, 4, 4, 5, 6}
	slice01 = append(slice01, 7) // el append aumenta el capacity * 2
	fmt.Println(slice01)

	slice02 := slice01[1:]
	slice03 := slice01[:len(slice01)-1] // [Desde donde : Cuantas posiciones]
	slice04 := slice01[:1]

	fmt.Println(slice02, slice03, slice04)

	//Maps
	mapa01 := map[uint]string{
		1: "Holi",
		2: "Chau",
		3: "",
	}

	var mapa02 map[uint]string
	mapa02 = make(map[uint]string)

	valueOrDefaultZero := mapa02[1]
	println(valueOrDefaultZero)
	value, ok := mapa02[1]
	if ok {
	}

	delete(mapa01, 1)

	//Custom Types
	// Toca verlos afuera

	// Enums

	//Functions

	myArray := []int{1, 2, 3}
	fmt.Println("___________________________________")
	AddNormal(myArray)
	Add(&myArray)
	fmt.Println(myArray)

	myMap := map[int]string{
		1: "Holi",
		2: "Chau",
	}
	AddKey(myMap)
	fmt.Println(myMap)
}

func Add(input *[]int) {
	(*input)[0] = 10

	newArr := append(*input, 23)
	*input = newArr
}

func AddNormal(input []int) {
	input[0] = 10

	input = append(input, 69)

	input[0] = 55
}

func AddKey(input map[int]string) {
	input[10] = "Dito"

	input = map[int]string{
		3: "Hola mundo",
	}

	input[10] = "Dito2"
}

// Custom types
// Type Alias
// No pueden asociarse metodos

type ID = uint64
type UUIDAlias = [32]byte

// types.ID

type UserIDs = []ID
type bytes = []byte

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

// Type Wrapper
// Si pueden asociarse metodos

type IDWrapper uint64

func (id IDWrapper) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

type UUID [32]byte

// abc12-asdas-asadasd
// 32 33 34

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
		// TODO: Dito quedo aquí
	}

	Struct struct {
		ID    uint64 `tag1:"tagValue1|tagValue2|tagValue3" tag2:"tagValue" tag3:"tagValue" gorm:"<-create;<-update;<-delete"`
		Value string `binding:"email"` // go-validator
	}

	Wrapper uint64

	WrapperFunc func()
)

// Ejercicio:
// Temas:
//	-> Variables - Condicionales - Switches - Casteo de Types (algo.(int))
//	-> Collectiones - Arrays - Slices - Maps
//	-> Tipos - Alias - Wrapper types
//	-> Funciones - Generics
//	-> Enumeradores
