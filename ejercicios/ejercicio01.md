# Bot de Trading de Forex

## Descripción

Desarrolla un bot de trading de Forex en Golang que analice las tasas de cambio de diferentes divisas a lo largo del
tiempo y decida si debe comprar, vender o no operar en base a un algoritmo simple.

## Datos de Entrada

- **Tasas de Cambio:**
    - **Divisa**: Símbolo de la divisa.
    - **Tasa**: Tasa de cambio en un momento dado.
    - **Fecha**: Fecha y hora de la tasa de cambio.

## Algoritmo de Decisión (opcional)

- **Comprar**: Si la tasa de cambio ha disminuido en un x% en comparación con la tasa anterior.
- **Vender**: Si la tasa de cambio ha aumentado en un x% en comparación con la tasa anterior.
- **No Operar**: Si la variación de la tasa no cumple con los criterios anteriores.

## Requisitos Matemáticos (opcional)

- **Cálculo de Variación de Tasa:** Se necesita calcular el cambio porcentual de la tasa de cambio utilizando la
  fórmula:

```math 
  \text{Variación} = \frac{\text{Tasa Actual} - \text{Tasa Anterior}}{\text{Tasa Anterior}} \times 100
```

### Enumeradores

Define un enumerador para las acciones del bot:

```go
package main

import (
	"fmt"
	"time"
)

type Action int

const (
	Buy Action = iota
	Sell
	Hold
)

func (a Action) String() string {
	return [...]string{"Buy", "Sell", "Hold"}[a]
}
```

### Colecciones y Estructuras de Datos

Define coleciones y estructuras de datos que te permitan almacenar los pares de divisas, y su cambio de precio en el
tiempo.
