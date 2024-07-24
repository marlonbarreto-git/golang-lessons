package main

var (
	globalVariable01 int
	globalVariable02 uint64 = 55555
	globalVariable03 any    = uint32(3)
	globalVariable04        = globalVariable03.(uint32)
	globalVariable05        = 1
)

func main() {
	// Variables
	var variable01 int
	var variable02 uint64 = 55555
	var variable03 any = uint32(3)
	var variable04 = variable03.(uint32)
	var variable05 = 1
	variable06 := 1

	// Constantes
	const constanteInt = 1
	const constanteUInt uint = 1

	_ = []any{
		variable01,
		variable02,
		variable03,
		variable04,
		variable05,
		variable06,
	}
}
