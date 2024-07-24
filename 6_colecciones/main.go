package main

import "fmt"

func main() {
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
		fmt.Println(value)
	}

	delete(mapa01, 1)
}
