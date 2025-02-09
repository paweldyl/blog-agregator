package main

import (
	"fmt"
	"gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	err = cfg.SetUser("Paweł")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg)
}
