package main

import (
    "io"
    "log"
    "net/http"
)

func ApiQuery(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "{\"status\": \"OK\"}")
}

func main() {
    http.HandleFunc("/", ApiQuery)
    err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
