//go:generate go run -mod=mod entc

//go:build ignore
// +build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	log.Println("Generating ent code...")
	if err := entc.Generate("./schema", &gen.Config{}); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
	log.Println("Ent code generation completed.")
}
