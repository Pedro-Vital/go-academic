package main

import "fmt"

func main() {
	notas := map[string][]float64{}
	// mapa de strings como chaves e slices como valores
	notas["Pedro"] = []float64{9.8, 5.32, 8.1}
	notas["Teo"] = []float64{10, 7.32, 4.75}
	fmt.Println(notas)
	fmt.Println(notas["Teo"])

	cursos := map[string][]string{
		"Teo":   {"ds", "estatistica", "python"},
		"Pedro": {"mecânica clássica", "termodinâmica"},
	}

	cursos["Ed"] = []string{"engenharia"}
	fmt.Println(cursos)
}
