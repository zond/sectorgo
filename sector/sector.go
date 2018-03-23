package main

import (
	"flag"
	"fmt"

	"github.com/zond/sectorgo"
)

func main() {
	userID := flag.String("userID", "", "Sector Alarm userID")
	password := flag.String("password", "", "Sector Alarm password")

	flag.Parse()

	if *userID == "" || *password == "" {
		flag.Usage()
		return
	}

	status, err := sectorgo.GetStatus(*userID, *password)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", status)
}
