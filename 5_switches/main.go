package main

func main() {
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
		println("Acciones por defecto")
	}

	var structValue struct{ ID uint64 }
	switch structValue {
	case struct{ ID uint64 }{
		ID: 0,
	}:
		println("hace match")
	}

	// Switch - Type
	var myVariable any = "tipo string"
	switch myValue := myVariable.(type) {
	case int, uint, float32, float64:
		println(myValue)
	case string:
		println(myValue)
	}
}
