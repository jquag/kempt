package cmd

import (
    "fmt"
    "github.com/jquag/kempt/conf"
    "github.com/jquag/kempt/util"
)

type Cmd struct {
    Name        string
    Usage       string
    NeedsConf   bool
    Handler     func(*conf.Config, string) error
    Description string
}

var Commands []Cmd

func init() {
    Commands = []Cmd{
        {
            "help", "help [SUBCMD]", false, runHelp,
            "Show help messages, use `help SUBCMD` for help on a specific subcommand",
        },
        {
            "init", "init [DIR=.]", false, runInit,
            fmt.Sprintf("Creates a default %s file in DIR\n%s", conf.ConfName, conf.Help()),
        },
        {
            "archive", "archive [DIR=.]", true, runArchive,
            fmt.Sprintf("Archives files in DIR according to %s\nArchiving may consist of moving files to the archive folders as well as zipping folders", conf.ConfName),
        },
        {
            "purge", "purge [DIR=.]", true, runPurge,
            fmt.Sprintf("Purges/deletes files in DIR and the archives according to %s", conf.ConfName),
        },
        {
            "run", "run [DIR=.]", true, runRun,
            fmt.Sprintf("Archives and purges files according to %s\nSee `help archive` and `help purge` for more info.", conf.ConfName),
        },
    }
}

func missingConfError(dir string) error {
    if dir == "." {
        return fmt.Errorf("%[1]s not found. Use `kempt init` to generate a default %[1]s file then edit as desired.", conf.ConfName)
    } else {
        return fmt.Errorf("%[1]s not found in '%[2]s'. Use `kempt init %[2]s` to generate a default %[1]s file then edit as desired.", conf.ConfName, dir)
    }
}

func Usage() {
    fmt.Println("Usage: kempt SUBCMD [args]")
    fmt.Println("Use `kempt help SUBCMD` for help on a specific subcommand")
    fmt.Println("Subcommands -")
    for _, cmd := range Commands {
        fmt.Printf("\t%s\n", cmd.Name)
    }
}

func Run(cmdName string, arg string) error {
    c := Lookup(cmdName)
    if c == nil {
        return fmt.Errorf("invalid subcommand: %s", cmdName)
    }

    if c.NeedsConf {
        arg := util.DefaultString(arg, ".")
        if !conf.Exists(arg) {
            return missingConfError(arg)
        } else {
            cfg, err := conf.Parse(arg)
            if err != nil {
                return err
            }
            return c.Handler(cfg, arg)
        }
    } else {
        return c.Handler(nil, arg)
    }
}

func Lookup(name string) *Cmd {
    for _, cmd := range Commands {
        if cmd.Name == name {
            return &cmd
        }
    }
    return nil
}

func runHelp(cfg *conf.Config, subcmd string) error {
    if subcmd == "" {
        fmt.Println("kempt, a utility for keeping a directory of files clean and kempt")
        fmt.Println(" - automatically moves files past a certain age into an archive, organized by date")
        fmt.Println(" - zips archive folders to save space")
        fmt.Println(" - purges (deletes) files of a certain age\n")
        Usage()
    } else {
        c := Lookup(subcmd)
        if c == nil {
            Usage()
        } else {
            util.PrintUnderlined(c.Name)
            fmt.Printf("\nUsage: kempt %s\n\n", c.Usage)
            fmt.Println(c.Description)
        }
    }
    return nil
}

func runInit(cfg *conf.Config, dir string) error {
    dir = util.DefaultString(dir, ".")
    if conf.Exists(dir) {
        confirmed, _ := util.Confirm(fmt.Sprintf("%s already exists. Do you want to overwrite it with the default configuration (yN)? ", conf.ConfName))
        if !confirmed {
            fmt.Println("aborting init")
            return nil
        }
    }
    return conf.Init(dir)
}

func runRun(cfg *conf.Config, dir string) (err error) {
    if err = runArchive(cfg, dir); err == nil {
        err = runPurge(cfg, dir)
    }
    return
}
