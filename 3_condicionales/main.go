package main

import (
	"fmt"
	"time"
)

func main() {
	// Forever -> While
	contador := 0
	for {
		fmt.Printf("While %d\n", contador)

		contador++
		if contador == 5 {
			break
		}
	}

	// For i traditional
	for i := uint(0); i <= 5; i++ {
		fmt.Printf("For i %d\n", i)
	}

	// For Range (index, value := range collection)
	for index, value := range []string{"holi", "chau"} {
		println(index, value)
	}

	// For Range Channels
	for value := range getChanWithData() {
		println(value)
	}
}

// getChanWithData nos entrega un canal en el que se recibiran elementos cada segundo,
// para este punto se puede ignorar puesto que se vera con mas detalle mas adelante
func getChanWithData() <-chan string {
	c := make(chan string)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			c <- fmt.Sprintf("Channel Item %d", i)
		}
		close(c)
	}()
	return c
}
