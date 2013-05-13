package main

import (
    "fmt"
    "regexp"
)

func main() {
    url := "/topic/cooperate/sadlove?fr=mp3_tuijian"
    rawRegex := `^/topic/(.*)/(.*)\?fr=(.*)$`
    routeRegex, _ := regexp.Compile(rawRegex)
    urlParams := routeRegex.FindStringSubmatch(url)[1:]
    fmt.Println(urlParams)
}
