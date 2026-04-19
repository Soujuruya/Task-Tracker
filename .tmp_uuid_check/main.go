package main

import (
  "fmt"
  "task-tracker-1/internal/pkg"
)

func main() {
  for i := 0; i < 3; i++ {
    id, err := pkg.GenerateID()
    if err != nil {
      panic(err)
    }
    fmt.Println(id)
  }
}
