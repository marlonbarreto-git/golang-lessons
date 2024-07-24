package main

import "fmt"

type Estado uint8

const (
	Inicial Estado = iota
	Procesando
	Aceptado
	Rechazado
)

var textoEstados = [...]string{
	Inicial:    "Inicial",
	Procesando: "Procesando",
	Aceptado:   "Aceptado",
	Rechazado:  "Rechazado",
}

func (e Estado) String() string {
	return textoEstados[e]
}

func (e Estado) SiguienteEstado(isValid ...bool) Estado {
	if e == Inicial {
		return Procesando
	}

	if len(isValid) > 0 && isValid[0] {
		return Aceptado
	}

	return Rechazado
}

func main() {
	fmt.Println(Inicial) // Imprime ´Inicial´ en lugar de 0 al haber implementado el metodo string
	fmt.Println(Procesando.SiguienteEstado(true))
	fmt.Println(Procesando.SiguienteEstado())
}
