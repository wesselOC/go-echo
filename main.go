package main

import (
    "math/rand"
    "fmt"
    "net/http"
    "strings"
    "log"
    "net"
    "io/ioutil"
    "encoding/json"
    "time"
    "bufio"
    "strconv"
    "os"
)

const (
    logFileName = "go-echo.log"
)

//Globals
var r *rand.Rand // Rand for this package.
var success bool = true


func init() {
    r = rand.New(rand.NewSource(time.Now().UnixNano()))
}


func sayhello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to IoP IP bot! Beedo boop bop beeda beep boop") // send data to client side
}


// FromRequest extracts the user IP address from req, if present.
func IPFromRequest(r *http.Request) (net.IP, error) {
    fmt.Println("Checking hosts...")
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
    }

    userIP := net.ParseIP(ip)
    if userIP == nil {
        return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
    }
    fmt.Println("IP found.." + userIP.String())
    return userIP, nil
}


func sayIp(w http.ResponseWriter, r *http.Request) {

    ip , err := IPFromRequest(r)

    if err != nil {
        log.Panic(err.Error())
        http.Error(w, err.Error(), 500)
        return
    }

    if(ip != nil){
        fmt.Fprintf(w, ip.String()) // send data to client side
    }else{
        fmt.Fprintf(w, "Not Found")
    }

}

// IsIPv4 check if the string is an IP version 4.
func IsIPv4(str string) bool {
    ip := net.ParseIP(str)
    return ip != nil && strings.Contains(str, ".")
}

// IsIPv6 check if the string is an IP version 6.
func IsIPv6(str string) bool {
    ip := net.ParseIP(str)
    return ip != nil && strings.Contains(str, ":")
}

func sayLocation(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()  // parse arguments, you have to call this by yourself

    var ip string = r.URL.Query().Get("ip");
    if ip == "" {
        http.Error(w, "IP not supplied", 400)
        return
    }

    //check that it is a valid ip
    if(!IsIPv4(ip) && !IsIPv6(ip) ){
        http.Error(w, "Invalid IP", 400)
        return
    }

    //cal thr location service
    response, err :=  http.Get("http://ipinfo.io/" + ip +  "/json")

    if err != nil {
        log.Fatal(err.Error())
        http.Error(w, err.Error(), 500)
        return
    } else {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err != nil {
            log.Fatal(err.Error())
            http.Error(w, err.Error(), 500)
            return
        }
        var dat map[string]interface{}
        if err := json.Unmarshal([]byte(contents), &dat); err != nil {
            log.Fatal(err.Error())
            http.Error(w, err.Error(), 500)
            return
        }
        fmt.Fprintf(w,dat["loc"].(string))
    }
}


func RandomString(strlen int) string {
    const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
    result := make([]byte, strlen)
    for i := range result {
        result[i] = chars[r.Intn(len(chars))]
    }
    return string(result)
}

func handleConnection(conn net.Conn, done chan bool) {

    fmt.Println("Handling new connection...")
    var randomString = RandomString(20)
    success = false

    //10 second time out
    timeoutDuration := 10 * time.Second
    conn.SetWriteDeadline(time.Now().Add(timeoutDuration))

    // send to socket
    fmt.Fprintf(conn, randomString+"\n")

    // Close connection when this function ends
    defer func() {
        fmt.Println("Closing connection...")
        conn.Close()
        done <- true
    }()

    bufReader := bufio.NewReader(conn)
    for {
        // Set a deadline for reading. Read operation will fail if no data
        // is received after deadline.
        conn.SetReadDeadline(time.Now().Add(timeoutDuration))
        // read in input from stdin
        message, err := bufReader.ReadString('\n')
        //received back the string

        if err != nil {
            fmt.Println(err)
            return
        }

        if(message == randomString + "\n"){
            success = true
            done <- true
            return
        }
    }
}

func checkConnection(w http.ResponseWriter, r *http.Request) {

    ip := r.URL.Query().Get("ip");
    if ip ==  "" {
        http.Error(w, "IP not supplied", 400)
        return
    }

    port := r.URL.Query().Get("port");
    if port ==  "" {
        http.Error(w, "port not supplied", 400)
        return
    }

    //check that it is a valid ip
    if(!IsIPv4(ip) && !IsIPv6(ip) ){
        http.Error(w, "Invalid IP", 400)
        return
    }

    // string to int
    intPort, err := strconv.Atoi(port)

    //check that it is a valid port
    if(err != nil  || intPort > 65535 || intPort <1 ){
        http.Error(w, "Invalid Port", 400)
        return
    }

    // Use this channel to follow the execution status
    // of our goroutines :D
    done := make(chan bool)

    for {
        // Get net.TCPConn object
        //10 second time out
        timeoutDuration := 10 * time.Second
        conn, err := net.DialTimeout("tcp", ip + ":" + port, timeoutDuration)

        //if there was an issue
        if err, ok := err.(net.Error); ok && err.Timeout() {
            fmt.Fprintf(w,"FAILED")
            return
        }

        if err != nil {
            fmt.Fprintf(w,"FAILED")
            return
        }
        go handleConnection(conn, done)

        <-done
        if(success){
            fmt.Fprintf(w,"OK")
        }else{
            fmt.Fprintf(w,"FAILED")
        }

        return

    }
}




func route(w http.ResponseWriter, r *http.Request) {

    //only allow GET
    if (r.Method != http.MethodGet){
        log.Fatal("ListenAndServe: ", http.ErrNotSupported)
        http.Error(w, "Method Not Allowed", 405)
        return
    }

    switch r.URL.Path {
        case "/":
            sayhello(w,r)
        case "/getip":
            sayIp(w,r)
        case "/location":
            sayLocation(w,r)
        case "/portcheck":
            checkConnection(w,r)
        default:
            http.Error(w, "Not found", 404)
            return
    }
}

func main() {

    log.SetFlags(log.Lshortfile)

    logf, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE,
        0640)
    if err != nil {
        log.Fatalln(err)
    }

    log.SetOutput(logf)

    http.HandleFunc("/", route) // set router
    err = http.ListenAndServe(":80", nil) // set listen port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}