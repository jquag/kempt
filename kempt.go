package main

import (
    "fmt"
    "github.com/jquag/kempt/cmd"
    "os"
)

func main() {
    if len(os.Args) < 2 || len(os.Args) > 3 {
        cmd.Usage()
        os.Exit(1)
    }

    subCommandName, extraArg := os.Args[1], ""
    if len(os.Args) == 3 {
        extraArg = os.Args[2]
    }

    if err := cmd.Run(subCommandName, extraArg); err != nil {
        fmt.Printf("Error! %s\n", err.Error())
        os.Exit(2)
    }
}
