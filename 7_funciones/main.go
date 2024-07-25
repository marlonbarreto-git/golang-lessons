package main

import (
	"fmt"
	"time"
)

func concatenar(inputA string, inputB string) (output string) {
	return inputA + inputB
}

func concatenarMultiple(input ...string) (outputA string, outputB int) {
	for i, s := range input {
		outputA += s
		outputB += i
	}

	return
}

func sumarConcatenar[T int | string](input ...T) (sum T) {
	for _, t := range input {
		sum += t
	}

	return
}

type Number interface {
	int | ~int
}

type NumeroPotente int

func (NumeroPotente) String() string {
	return fmt.Sprint("Potente")
}

func imprimir[T any](item T) {
	fmt.Println(item)
}

func imprimirNumero[T Number](item T) {
	fmt.Println(item)
}

func MapArray[Input any, Output any](arrayA []Input, fn func(Input) Output) (arrayOut []Output) {
	for _, input := range arrayA {
		mappedOutput := fn(input)
		arrayOut = append(arrayOut, mappedOutput)
	}

	return
}

func medirDuracion[T any](fn func() T) (T, time.Duration) {
	tiempoDeInicio := time.Now().UnixNano()
	resultado := fn()
	tiempoFinal := time.Now().UnixNano()

	return resultado, time.Duration(tiempoFinal - tiempoDeInicio)
}

func ConnectDB() (db string, closeConn func()) {
	var err error

	return "database", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Closing connection ...", err)
	}
}

func potencia(base float64, exponente float64) (resultado float64) {
	if exponente == 0 && base == 0 {
		panic("indeterminado")
	}

	if exponente == 0 {
		return 1
	}

	if base == 0 {
		return 0
	}

	if exponente < 0 {
		return 1 / potencia(base, -exponente)
	}

	return base * potencia(base, exponente-1)
}

func main() {
	holaMundo := concatenar("Hola ", "Mundo")
	fmt.Println(holaMundo)

	ejemplo := []string{
		"Go ",
		"o tambien conocido como Golang ",
		"es un lenguaje de programacion",
	}

	texto, numeroDeCadenas := concatenarMultiple()
	fmt.Printf("Cadenas: %d\nTexto: %s\n", numeroDeCadenas, texto)

	texto, numeroDeCadenas = concatenarMultiple(ejemplo...)
	fmt.Printf("Cadenas: %d\nTexto: %s\n", numeroDeCadenas, texto)

	imprimir[uint64](1)
	imprimir("Hola Dito")
	imprimir(map[string]string{
		"Hola": "Mundo",
	})

	db, closeFunc := ConnectDB()
	fmt.Println(db)
	closeFunc()

	var miNumeroPotente NumeroPotente = 10

	imprimirNumero(miNumeroPotente)

	texto2 := sumarConcatenar(ejemplo...)
	fmt.Printf("Texto: %s\n", texto2)

	numero := sumarConcatenar(10, 20, 30)
	fmt.Printf("numero: %d\n", numero)

	resultado, duracion := medirDuracion(func() string {
		time.Sleep(1 * time.Millisecond)
		return ejemplo[0]
	})
	fmt.Printf("Duracion: %v\nResultado: %v\n\n", duracion, resultado)

	fmt.Printf("Potencia: [")
	for i := 0; i < 10; i++ {
		potenciacion := potencia(0.2, float64(-i))
		fmt.Printf(" %.f", potenciacion)
	}
	fmt.Printf(" ]\n")

}
