package main

import "fmt"

type Role string

const (
	Unknown   Role = ""
	Guest     Role = "guest"
	Member    Role = "member"
	Moderator Role = "moderator"
	Admin     Role = "admin"
)

func main2() {
	fmt.Printf("%c", byte(7))

	r := Unknown

	switch r {
	case Unknown:
		fmt.Println("good")
	case Guest:
		fmt.Print("bad")
	}
}
