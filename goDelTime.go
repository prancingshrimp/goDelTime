package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    // "sort"
    "runtime"
    "strconv"
    "strings"
)

//Name: processor0 Content: [0.01 0.02]
type procDir struct {
    Name    string
    Content []string
}

func main() {
    fmt.Println("goDelTime v0.01 2013-04-02")
    fileName := "timeDirs"
    notRemoveSlice := readFile(fileName)

    currentDir := "."
    search := "processor"
    procNames := getDirNames(currentDir, search, "")
    if len(procNames) == 0 {
        fmt.Println("There are no " + search + " directories!")
        os.Exit(1)
    }
    fmt.Println("-> Found " + strconv.Itoa(len(procNames)) +
        " " + search + " directories.\n")
    //fmt.Println(procNames)

    //Slice of procDirs: every processor with individual content
    procTimes := []procDir{}
    for _, item := range procNames {
        procContent := getDirNames(item, "", "constant")
        procTimes = append(procTimes, procDir{item, procContent})
    }
    //search for time dirs which are in every processor dir
    commonTimes, procTimes := getCommonTimes(procTimes)
    // fmt.Println(commonTimes)
    // fmt.Println(procTimes)

    //substract the time dirs which will not be deleted
    commonTimes, notRemoveSlice = excludeTimeDirs(commonTimes, notRemoveSlice)
    fmt.Println("Found the following common time directories for deletion:")
    fmt.Println(commonTimes)

    //User choose time dirs for deletion
    removeSlice, removeSliceProc, notRemoveSlice :=
        askWhatToRemove(commonTimes, notRemoveSlice, procTimes)
    if len(removeSlice) == 0 && len(removeSliceProc) == 0 {
        writeTimeDirs(fileName, notRemoveSlice)
        fmt.Println("Nothing to delete!")
        os.Exit(0)
    }
    fmt.Println("Select the following time dirs for deletion:")
    fmt.Println(removeSlice)
    fmt.Println(removeSliceProc)

    //Final question: delete them or not -> you have to say yes or no
    for {
        inputReader := bufio.NewReader(os.Stdin)
        fmt.Println("Do you want to delete them? [y/n]() ")
        input, _ := inputReader.ReadString('\n')
        input = strings.Trim(strings.Split(input, " ")[0], "\n")
        if input == "y" || input == "yes" || input == "Y" {
            removeTimeDirs(removeSlice, removeSliceProc, procNames)
            writeTimeDirs(fileName, notRemoveSlice)
            fmt.Println("Done!")
            os.Exit(0)
        } else if input == "n" || input == "no" || input == "N" {
            writeTimeDirs(fileName, notRemoveSlice)
            fmt.Println("Done!")
            os.Exit(0)
        }
    }

}

func writeTimeDirs(fileName string, notRemoveSlice []string) {
    // Write time dirs into timeDirs file which are not removed

    // os.RemoveAll("./" + fileName)
    output, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        panic("Cannot write to " + fileName + "!")
    }
    defer output.Close()
    writer := bufio.NewWriter(output)
    for _, item := range notRemoveSlice {
        writer.WriteString(item + "\n")
    }
    writer.Flush()
}

func removeTimeDirs(removeSlice []string, removeSliceProc []string,
    procNames []string) {
    //Remove selected time dirs the concurrent way with channels
    runtime.GOMAXPROCS(runtime.NumCPU())

    channel := make(chan bool)
    procNamesNum := len(procNames)

    for i := 0; i < procNamesNum; i++ {
        go removeInProc(procNames[i], removeSlice, channel)
    }
    //Wait until all go functions stopped
    for i := 0; i < procNamesNum; i++ {
        <-channel
    }
    //Remove non-common time dirs not concurrent
    for _, item := range removeSliceProc {
        os.RemoveAll("./" + item)
        fmt.Printf("remove -> %s\n", item)
    }
}

func removeInProc(name string, removeSlice []string, channel chan bool) {
    //Concurrent remove of time dirs in processor dirs
    for _, item := range removeSlice {
        os.RemoveAll("./" + name + "/" + item)
        fmt.Printf("%s: remove -> %s\n", name, item)
    }
    channel <- true
}

func askWhatToRemove(commonTimes []string, notRemoveSlice []string,
    procTimes []procDir) ([]string, []string, []string) {
    //Ask for common time dirs
    fmt.Println()
    removeSlice := []string{}
    for _, item := range commonTimes {
        inputReader := bufio.NewReader(os.Stdin)
        fmt.Printf("Delete %s? [y/n](n) ", item)
        input, _ := inputReader.ReadString('\n')
        input = strings.Trim(strings.Split(input, " ")[0], "\n")
        if input != "y" {
            notRemoveSlice = append(notRemoveSlice, item)
        } else {
            removeSlice = append(removeSlice, item)
        }
    }
    fmt.Println()

    //Ask for non-common time dirs
    removeSliceProc := []string{}
    for _, item := range procTimes {
        if len(item.Content) != 0 {
            for i, _ := range item.Content {
                inputReader := bufio.NewReader(os.Stdin)
                fmt.Printf("Delete %s/%s? [y/n](n) ", item.Name,
                    item.Content[i])
                input, _ := inputReader.ReadString('\n')
                input = strings.Trim(strings.Split(input, " ")[0], "\n")
                if input == "y" {
                    removeSliceProc = append(removeSliceProc,
                        item.Name+"/"+item.Content[i])
                }
            }
        }
    }
    fmt.Println()

    return removeSlice, removeSliceProc, notRemoveSlice
}

func excludeTimeDirs(commonTimes []string, notRemoveSlice []string) ([]string,
    []string) {
    //Substract time dirs which will be not removed but only if
    //there is something to substract
    if len(notRemoveSlice) != 0 {
        inputReader := bufio.NewReader(os.Stdin)
        fmt.Println("Want use timeDirs? [y/n](y)")
        input, _ := inputReader.ReadString('\n')
        input = strings.Trim(strings.Split(input, " ")[0], "\n")
        if input != "n" && input != "N" && input != "no" && input != "No" {
            for _, item := range notRemoveSlice {
                index := searchInSlice(commonTimes, item)
                if index != -1 {
                    //fmt.Println(index)
                    commonTimes = append(commonTimes[:index],
                        commonTimes[index+1:]...)
                    //fmt.Println(commonTimes)
                }
            }
        } else {
            notRemoveSlice = []string{}
        }
    }
    return commonTimes, notRemoveSlice
}

func getCommonTimes(procTimes []procDir) ([]string, []procDir) {
    //Found out which time dirs are in every processor dir

    // fmt.Println(procTimes)
    procNumber := len(procTimes)
    commonTimes := []string{}
    for _, item := range procTimes[0].Content {
        count := 1
        for i := 1; i < procNumber; i++ {
            for _, cprItem := range procTimes[i].Content {
                if item == cprItem {
                    count++
                }
            }
        }
        if count == procNumber {
            commonTimes = append(commonTimes, item)
        }
    }

    for i := 0; i < procNumber; i++ {
        for _, item := range commonTimes {
            index := searchInSlice(procTimes[i].Content, item)
            procTimes[i].Content = append(procTimes[i].Content[:index],
                procTimes[i].Content[index+1:]...)
        }
    }

    return commonTimes, procTimes
}

func searchInSlice(strSlice []string, searchStr string) int {
    //Linear search function (not fast) which returns the index of 
    //the searched item or -1 if it is not in the slice

    // sort.Strings(strSlice)
    // return sort.SearchStrings(strSlice, searchStr)
    for i, item := range strSlice {
        if item == searchStr {
            return i
        }
    }
    return -1
}

func insertString(insert string, slice []string, pos uint16) []string {
    //Insert a string at a certain position but not used at the moment

    //strSlice = insertString("_", strSlice, 0)
    return append(slice[:pos], append([]string{insert}, slice[pos:]...)...)
}

func readFile(fileName string) []string {
    //Reads the timeDir file in bytes format and
    //converts the data into strings

    rawBytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        fmt.Println("Cannot find " + fileName + " file -> create a new one!")
    }
    text := strings.Trim(string(rawBytes), "\n")

    if len(text) == 0 {
        return []string{}
    }
    notRemoveSlice := strings.Split(text, "\n")
    fmt.Println("\nRead timeDirs (" + strconv.Itoa(len(notRemoveSlice)) +
        " files): ")
    fmt.Println(notRemoveSlice)
    return notRemoveSlice
}

func getDirNames(dir string, containStr string, noDir string) []string {
    //Look for processor dirs, quit the program
    //if there are no processor dirs

    fd, err := os.Open(dir)
    if err != nil {
        panic("Cannot open " + dir + "!")
    }
    defer fd.Close()

    dirContent, err := fd.Readdir(-1)
    if err != nil {
        panic("Cannot read the directory!")
    }

    dirNames := []string{}
    for _, item := range dirContent {
        itemName := item.Name()
        itemPath := "./" + dir + "/" + itemName
        itemStat, err := os.Stat(itemPath)
        if err != nil {
            panic("Cannot get the status of" + itemPath + "!")
        }
        if itemStat.IsDir() {
            if strings.Contains(itemName, containStr) {
                if itemName != noDir {
                    dirNames = append(dirNames, itemName)
                }
            }
        }
    }
    return dirNames
}
