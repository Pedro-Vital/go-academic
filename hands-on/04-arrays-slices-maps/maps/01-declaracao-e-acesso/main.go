package main

import "fmt"

func main() {

	idades := make(map[string]int) // declaração
	idades["Pedro"] = 25
	idades["Teo"] = 33

	fmt.Println("Idades:", idades)

	alturas := map[string]float64{} // declaração mais comum
	alturas["Pedro"] = 1.76
	fmt.Println("Alturas:", alturas)

	alturas["Teo"] = 1.82
	alturaTeo, ok := alturas["Teo"]
	if ok {
		fmt.Println("Altura Teo:", alturaTeo, "ok:", ok)
	} else {
		fmt.Println("Não encontrei")
	}

	// Ou podemos fazer da forma a seguir:
	if alturaTeo, ok := alturas["Teo"]; ok {
		fmt.Println("Altura Teo:", alturaTeo)
	} else {
		fmt.Println("Não encontrei")
	}
	// Aqui, alturaTeo e ok existem apenas dentro da estrutura condicional

}
