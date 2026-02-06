// Package main - Chapter 068: Reflection
// El paquete reflect permite inspeccionar y manipular tipos en runtime.
// Poderoso pero usar con cuidado - preferir alternativas cuando sea posible.
package main

import (
	"fmt"
	"reflect"
)

type User struct {
	ID        int    `json:"id" validate:"required"`
	Name      string `json:"name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Age       int    `json:"age,omitempty"`
	isPrivate bool   // No exportado
}

func (u User) Greet() string {
	return "Hello, " + u.Name
}

func (u *User) SetName(name string) {
	u.Name = name
}

func main() {
	fmt.Println("=== REFLECTION EN GO ===")

	// ============================================
	// CONCEPTOS BÁSICOS
	// ============================================
	fmt.Println("\n--- Conceptos Básicos ---")

	// reflect.TypeOf - obtener tipo
	// reflect.ValueOf - obtener valor

	var x int = 42
	t := reflect.TypeOf(x)
	v := reflect.ValueOf(x)

	fmt.Printf("Type: %v\n", t)           // int
	fmt.Printf("Kind: %v\n", t.Kind())    // int
	fmt.Printf("Value: %v\n", v)          // 42
	fmt.Printf("Interface: %v\n", v.Interface()) // 42

	// ============================================
	// KINDS
	// ============================================
	fmt.Println("\n--- Kinds ---")

	values := []any{42, "hello", 3.14, true, []int{1, 2}, map[string]int{}, User{}}

	for _, val := range values {
		t := reflect.TypeOf(val)
		fmt.Printf("Value: %-15v Type: %-15v Kind: %v\n", val, t, t.Kind())
	}

	// ============================================
	// INSPECCIONAR STRUCTS
	// ============================================
	fmt.Println("\n--- Inspeccionar Structs ---")

	user := User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
	userType := reflect.TypeOf(user)
	userValue := reflect.ValueOf(user)

	fmt.Printf("Type: %v\n", userType.Name())
	fmt.Printf("NumField: %d\n", userType.NumField())

	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		value := userValue.Field(i)

		// Verificar si el campo es exportado
		if field.IsExported() {
			fmt.Printf("  %s: %v (type: %v, tag: %v)\n",
				field.Name, value.Interface(), field.Type, field.Tag)
		} else {
			fmt.Printf("  %s: [unexported] (type: %v)\n",
				field.Name, field.Type)
		}
	}

	// ============================================
	// TAGS DE STRUCT
	// ============================================
	fmt.Println("\n--- Struct Tags ---")

	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)

		jsonTag := field.Tag.Get("json")
		validateTag := field.Tag.Get("validate")

		if jsonTag != "" || validateTag != "" {
			fmt.Printf("Field: %s\n", field.Name)
			fmt.Printf("  json: %q\n", jsonTag)
			fmt.Printf("  validate: %q\n", validateTag)
		}
	}

	// ============================================
	// INSPECCIONAR MÉTODOS
	// ============================================
	fmt.Println("\n--- Inspeccionar Métodos ---")

	// Para métodos con value receiver
	userType = reflect.TypeOf(user)
	fmt.Printf("Methods on User (value):\n")
	for i := 0; i < userType.NumMethod(); i++ {
		method := userType.Method(i)
		fmt.Printf("  %s: %v\n", method.Name, method.Type)
	}

	// Para métodos con pointer receiver
	userPtrType := reflect.TypeOf(&user)
	fmt.Printf("Methods on *User (pointer):\n")
	for i := 0; i < userPtrType.NumMethod(); i++ {
		method := userPtrType.Method(i)
		fmt.Printf("  %s: %v\n", method.Name, method.Type)
	}

	// ============================================
	// MODIFICAR VALORES
	// ============================================
	fmt.Println("\n--- Modificar Valores ---")

	// Para modificar, necesitamos un puntero
	x = 100
	xPtr := reflect.ValueOf(&x)
	xVal := xPtr.Elem() // Obtener el valor apuntado

	if xVal.CanSet() {
		xVal.SetInt(200)
	}
	fmt.Printf("Modified x: %d\n", x)

	// Modificar campos de struct
	userPtr := reflect.ValueOf(&user)
	userElem := userPtr.Elem()

	nameField := userElem.FieldByName("Name")
	if nameField.CanSet() {
		nameField.SetString("Bob")
	}
	fmt.Printf("Modified user: %+v\n", user)

	// ============================================
	// CREAR VALORES DINÁMICAMENTE
	// ============================================
	fmt.Println("\n--- Crear Valores Dinámicamente ---")

	// Crear slice
	sliceType := reflect.SliceOf(reflect.TypeOf(0))
	slice := reflect.MakeSlice(sliceType, 0, 10)
	slice = reflect.Append(slice, reflect.ValueOf(1))
	slice = reflect.Append(slice, reflect.ValueOf(2))
	slice = reflect.Append(slice, reflect.ValueOf(3))
	fmt.Printf("Created slice: %v\n", slice.Interface())

	// Crear map
	mapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0))
	m := reflect.MakeMap(mapType)
	m.SetMapIndex(reflect.ValueOf("one"), reflect.ValueOf(1))
	m.SetMapIndex(reflect.ValueOf("two"), reflect.ValueOf(2))
	fmt.Printf("Created map: %v\n", m.Interface())

	// Crear struct
	newUser := reflect.New(userType).Elem()
	newUser.FieldByName("ID").SetInt(99)
	newUser.FieldByName("Name").SetString("NewUser")
	newUser.FieldByName("Email").SetString("new@example.com")
	fmt.Printf("Created struct: %+v\n", newUser.Interface())

	// ============================================
	// LLAMAR FUNCIONES
	// ============================================
	fmt.Println("\n--- Llamar Funciones ---")

	// Llamar método
	greetMethod := reflect.ValueOf(user).MethodByName("Greet")
	results := greetMethod.Call(nil)
	fmt.Printf("Greet() returned: %v\n", results[0].Interface())

	// Llamar función con argumentos
	add := func(a, b int) int { return a + b }
	addValue := reflect.ValueOf(add)
	args := []reflect.Value{reflect.ValueOf(3), reflect.ValueOf(4)}
	result := addValue.Call(args)
	fmt.Printf("add(3, 4) = %v\n", result[0].Interface())

	// ============================================
	// TYPE SWITCH VS REFLECTION
	// ============================================
	fmt.Println("\n--- Type Switch vs Reflection ---")

	printType := func(v any) {
		// Preferir type switch cuando sea posible
		switch val := v.(type) {
		case int:
			fmt.Printf("int: %d\n", val)
		case string:
			fmt.Printf("string: %s\n", val)
		case bool:
			fmt.Printf("bool: %v\n", val)
		default:
			// Solo usar reflection cuando necesario
			fmt.Printf("unknown: %v (type: %T)\n", val, val)
		}
	}

	printType(42)
	printType("hello")
	printType(User{Name: "Test"})

	// ============================================
	// DEEP EQUAL
	// ============================================
	fmt.Println("\n--- DeepEqual ---")

	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	slice3 := []int{1, 2, 4}

	fmt.Printf("slice1 == slice2: %v\n", reflect.DeepEqual(slice1, slice2))
	fmt.Printf("slice1 == slice3: %v\n", reflect.DeepEqual(slice1, slice3))

	map1 := map[string]int{"a": 1, "b": 2}
	map2 := map[string]int{"a": 1, "b": 2}
	fmt.Printf("map1 == map2: %v\n", reflect.DeepEqual(map1, map2))

	// ============================================
	// CASOS DE USO VÁLIDOS
	// ============================================
	fmt.Println("\n--- Casos de Uso Válidos ---")
	fmt.Println(`
CUÁNDO USAR REFLECTION:

✓ Serialización (JSON, XML, etc.)
✓ ORM/Database mapping
✓ Dependency injection frameworks
✓ Validadores genéricos
✓ Testing (comparar structs)
✓ Configuración dinámica

CUÁNDO NO USAR REFLECTION:

✗ Cuando hay alternativa con generics
✗ Para código crítico en rendimiento
✗ Cuando type assertions bastan
✗ Para operaciones simples en tipos conocidos

ALTERNATIVAS:

1. Type switch:
   switch v := x.(type) { ... }

2. Generics (Go 1.18+):
   func Process[T any](items []T) { ... }

3. Interfaces:
   type Processor interface { Process() }`)
	// ============================================
	// EJEMPLO: VALIDADOR SIMPLE
	// ============================================
	fmt.Println("\n--- Ejemplo: Validador ---")

	validate := func(v any) []string {
		var errors []string
		val := reflect.ValueOf(v)
		typ := reflect.TypeOf(v)

		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			typ = typ.Elem()
		}

		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := val.Field(i)

			tag := field.Tag.Get("validate")
			if tag == "" {
				continue
			}

			// Validación simple de "required"
			if tag == "required" || contains(tag, "required") {
				if isZero(fieldVal) {
					errors = append(errors, fmt.Sprintf("%s is required", field.Name))
				}
			}
		}
		return errors
	}

	invalidUser := User{ID: 0, Name: "", Email: ""}
	errs := validate(invalidUser)
	fmt.Printf("Validation errors: %v\n", errs)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}

func isZero(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

/*
RESUMEN DE REFLECTION:

OBTENER TIPO Y VALOR:
t := reflect.TypeOf(x)
v := reflect.ValueOf(x)

KINDS COMUNES:
reflect.Int, reflect.String, reflect.Bool
reflect.Slice, reflect.Map, reflect.Struct
reflect.Ptr, reflect.Interface, reflect.Func

INSPECCIONAR STRUCTS:
t.NumField()
t.Field(i).Name
t.Field(i).Type
t.Field(i).Tag.Get("json")

MODIFICAR VALORES:
v := reflect.ValueOf(&x).Elem()
if v.CanSet() {
    v.SetInt(42)
}

CREAR VALORES:
reflect.MakeSlice(type, len, cap)
reflect.MakeMap(type)
reflect.New(type)

LLAMAR FUNCIONES:
method := v.MethodByName("Method")
results := method.Call(args)

COMPARAR:
reflect.DeepEqual(a, b)

REGLAS:
1. Preferir type switch sobre reflection
2. Preferir generics sobre reflection
3. Solo usar cuando realmente necesario
4. Reflection es lento - evitar en hot paths
5. Código con reflection es más difícil de mantener
*/
