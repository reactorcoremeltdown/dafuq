package main

import (
    "fmt"
    "time"
    "os"
    "os/exec"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "strconv"
    "bytes"
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
    Output string
    Counter int
    Status int
    CurrentStatus int
}

var configArray []config

func logErr(desc string, e error) {
    if e != nil {
        log.Println(desc + ": " + e.Error())
    }
}

func encodeConfig(res http.ResponseWriter, req *http.Request) {
    configJson, err := json.Marshal(configArray)
    logErr("Cannot encode to JSON", err)
    fmt.Fprint(res, string(configJson))
}

func writeStateFile(path string) (error) {
    configJson, err := json.Marshal(configArray)
    if err != nil {
        return err
    }
    data := []byte(configJson)
    err = ioutil.WriteFile(path, data, 0644)
    if err != nil {
        return err
    }

    return nil
}

func loadState(saved, loaded []config) ([]config){
    for _, key_loaded := range(loaded) {
        //fmt.Println("Loaded key: " + key_loaded.Name)
        for _, key_saved := range(saved) {
            //fmt.Println("Saved key: " + key_saved.Name)
            if key_loaded.Name == key_saved.Name {
                key_loaded.Counter = key_saved.Counter
                key_loaded.Status = key_saved.Status
                key_loaded.CurrentStatus = key_saved.CurrentStatus
            }
        }
    }

    log.Println("Loading state completed")
    return loaded
}

func main() {
    var configPath string
    configPathFromEnv, configPathFromEnvPresent := os.LookupEnv("CONFIG_PATH")
    if configPathFromEnvPresent {
        configPath = configPathFromEnv
    } else {
        configPath = "/etc/dafuq/config.ini"
    }
    cfg, err := ini.Load(configPath)
    if err != nil {
        fmt.Printf("Failed to load config file: %v", err)
        os.Exit(1)
    }

    configsDir := cfg.Section("main").Key("configs").String()
    pluginsDir := cfg.Section("main").Key("plugins").String()
    notifiersDir := cfg.Section("main").Key("notifiers").String()
    stateFilePath := cfg.Section("main").Key("stateFile").String()
    address := cfg.Section("main").Key("address").String()
    port := cfg.Section("main").Key("port").String()

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
        } else {
            log.Println("Loaded config file: " + configsDir + "/" + configFile.Name())
        }

        container.Name = configIni.Section("config").Key("name").String()
        container.Description = configIni.Section("config").Key("description").String()
        container.Plugin = configIni.Section("config").Key("plugin").String()
        container.Argument = configIni.Section("config").Key("argument").String()
        interval, _ := time.ParseDuration(configIni.Section("config").Key("interval").String())
        seconds := int(interval.Seconds())
        if seconds < 5 {
            container.Interval = 5
        } else {
            container.Interval = seconds
        }
        container.Notify = configIni.Section("config").Key("notify").String()
        container.Output = "Waiting for output"
        container.Counter = 0
        container.Status = 0
        container.CurrentStatus = 0

        configArray = append(configArray, container)
    }

    loadedState := make([]config,0)
    stateData, err := ioutil.ReadFile(stateFilePath)
    if err != nil {
        log.Println("Unable to load state from file: " + err.Error())
    } else {
        err = json.Unmarshal(stateData, &loadedState)
        if err != nil {
            log.Println("Unable to decode JSON from state data: " + err.Error())
        } else {
            configArray = loadState(loadedState, configArray)
        }
    }


    go func(){
        http.HandleFunc("/", encodeConfig)
        log.Println(http.ListenAndServe(address + ":" + port, nil))
    }()
    for {
        for index, _ := range configArray {
            configArray[index].Counter = configArray[index].Counter + 1
            if (configArray[index].Counter == configArray[index].Interval) {
                go func(i int){
                    log.Println("Running check: " + configArray[i].Name)
                    cmd := exec.Command("/bin/sh", "-c", pluginsDir + "/" + configArray[i].Plugin + " " + configArray[i].Argument)
                    var outputBuffer bytes.Buffer
                    cmd.Stdout = &outputBuffer
                    cmd.Env = os.Environ()
                    cmd.Env = append(cmd.Env,
                                        "PLUGINSDIR=" + pluginsDir)
                    if err := cmd.Run() ; err != nil {
                        if exitError, ok := err.(*exec.ExitError); ok {
                            configArray[i].CurrentStatus = exitError.ExitCode()
                        }
                    } else {
                        configArray[i].CurrentStatus = 0
                    }
                    configArray[i].Output = outputBuffer.String()

                    if configArray[i].CurrentStatus != configArray[i].Status {
                        err := writeStateFile(stateFilePath)
                        logErr("Unable to write data to state file", err)
                        log.Println("Status of check " + 
                                    configArray[i].Name + 
                                    " changed from " + 
                                    strconv.Itoa(configArray[i].Status) +
                                    " to " +
                                    strconv.Itoa(configArray[i].CurrentStatus))
                        alert := exec.Command("/bin/sh", "-c", notifiersDir + "/" + configArray[i].Notify)
                        alert.Env = os.Environ()
                        alert.Env = append(alert.Env,
                                            "NAME=" + configArray[i].Name,
                                            "STATUS=" + strconv.Itoa(configArray[i].CurrentStatus),
                                            "DESCRIPTION=" + configArray[i].Description,
                                            "MESSAGE=" + outputBuffer.String())
                        err = alert.Run()
                        log.Println("Unable to launch alert", err)
                    }
                    configArray[i].Status = configArray[i].CurrentStatus
                }(index)
                configArray[index].Counter = 0
            }
        }
        time.Sleep(1 * time.Second)
    }
}
