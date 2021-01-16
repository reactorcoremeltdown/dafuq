package main

import (
    "fmt"
    "time"
    "os"
    "io/ioutil"
    "log"

    "gopkg.in/ini.v1"
)

type config struct {
    Name string
    Plugin string
    Argument string
    Interval int
    Description string
    Notify string
    Counter int
    Status int
}

var configArray []config

func main() {
    cfg, err := ini.Load("shovel.ini")
    if err != nil {
        fmt.Printf("Failed to load config file: %v", err)
        os.Exit(1)
    }

    configsDir := cfg.Section("main").Key("configs").String()
    pluginsDir := cfg.Section("main").Key("plugins").String()
    notifiersDir := cfg.Section("main").Key("notifiers").String()

    configFiles, err := ioutil.ReadDir(configsDir + "/")
    if err != nil {
        fmt.Printf("Failed to read directory contents: %v", err)
        os.Exit(1)
    }

    for _, configFile := range configFiles {
        var container config
        configIni, err := ini.Load(configsDir + "/" + configFile.Name())
        if err != nil {
            log.Println("Failed to parse config file: " + err.Error())
        }

        container.Name = configIni.Section("config").Key("name").String()
        container.Description = configIni.Section("config").Key("description").String()
        container.Plugin = configIni.Section("config").Key("plugin").String()
        container.Argument = configIni.Section("config").Key("argument").String()
        interval, _ := time.ParseDuration(configIni.Section("config").Key("interval").String())
        container.Interval = int(interval.Seconds())
        container.Notify = configIni.Section("config").Key("notify").String()
        container.Counter = 0
        container.Status = 0

        configArray = append(configArray, container)
    }
    for {
        for index, _ := range configArray {
            configArray[index].Counter = configArray[index].Counter + 1
            if (configArray[index].Counter == configArray[index].Interval) {
                go func(name, command, argument, notifier string){
                    log.Println("Running check: " + name)
                }(configArray[index].Name,
                    pluginsDir + "/" + configArray[index].Plugin,
                    configArray[index].Argument,
                    notifiersDir + "/" + configArray[index].Notify)
                configArray[index].Counter = 0
            }
        }
        time.Sleep(1 * time.Second)
    }
}
