package main

import "fmt"

func main() {
	// Condicion basica
	if true {
		fmt.Println("se debe imprimit esto")
	}

	if false {
		fmt.Println("nunca se imprimira esto")
	}

	// Multiples condiciones
	condicion01 := 1 == 1
	condicion02 := 1 == 2

	if condicion01 {
		fmt.Println("se cumple la condicion 1")
	} else if condicion02 {
		fmt.Println("se cumle la condicion 2")
	} else {
		fmt.Println("no se cumple ninguna de las condiciones")
	}

	// Condicional ternario
	defaultValue := 2
	if condicion01 {
		defaultValue = 1
	}

	fmt.Println(defaultValue)
}
