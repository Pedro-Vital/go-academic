package main

import (
	"fmt"
	"os"
)

func main() {
	tiposMap := map[string]float64{
		"casquinha": 1.00,
		"cascão":    2.00,
		"cestinha":  3.00,
	}

	saboresMap := map[string]float64{
		"morango":   0.1,
		"creme":     0.2,
		"chocolate": 0.3,
	}

	coberturasMap := map[string]float64{
		"caramelo":  0.01,
		"morango":   0.02,
		"chocolate": 0.03,
	}

	items := map[string]map[string]float64{ // mapa com mapas como valores
		"tipos":      tiposMap,
		"sabores":    saboresMap,
		"coberturas": coberturasMap,
	}

	var tipo, sabor, cobertura string
	var total float64

	fmt.Printf("Escolha um tipo [casquinha/cascão/cestinha]: ")
	fmt.Scanf("%s", &tipo)

	if valor, ok := items["tipos"][tipo]; !ok {
		fmt.Println("Tipo inválido")
		os.Exit(1) // sai do programa
	} else {
		total += valor
	}

	fmt.Printf("Escolha o sabor [morango/creme/chocolate]: ")
	fmt.Scanf("%s", &sabor)

	if valor, ok := items["sabores"][sabor]; !ok {
		fmt.Println("sabor inválido")
		os.Exit(1) // sai do programa
	} else {
		total += valor
	}

	fmt.Printf("Escolha um cobertura [caramelo/morango/chocolate]: ")
	fmt.Scanf("%s", &cobertura)

	if valor, ok := items["coberturas"][cobertura]; !ok {
		fmt.Println("Cobertura inválida")
		os.Exit(1) // sai do programa
	} else {
		total += valor
	}

	fmt.Printf("\nTotal: R$%.2f\n\n", total)

}
