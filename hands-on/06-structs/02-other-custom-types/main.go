package main

import "fmt"

type str string

func (text str) log() {
	fmt.Println(text)
}

type Altura float64
type Peso float64

func IMC(altura Altura, peso Peso) float64 {
	return float64(peso) / float64(altura * altura)
}

func main() {
	altura := Altura(1.80)
	peso := Peso(75.0)
	imc := IMC(altura, peso)
	fmt.Println("IMC:", imc)

	text := str("Hello, World!")
	text.log()
}
