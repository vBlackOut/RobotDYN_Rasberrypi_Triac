package main

import (
  "fmt"
  "os"
  //"os/signal"
  "syscall"
  "time"
  "bufio"
  "strconv"
  "github.com/warthog618/gpiod"
  "github.com/warthog618/gpiod/device/rpi"
  "strings"
  "sync"
)
var percent1 = uint32(0)
var percent2 = uint32(0)
var c, errchip = gpiod.NewChip("gpiochip0")

func getInt(key string) uint32 {
    //s := os.Getenv(key)
    if signals, err := strconv.ParseFloat(key, 32); err == nil {
      if (signals >= 90) && (signals <= 100) {
	signals = 9000-(signals*80)
      } else if signals <= 100 {
        signals = 9000-(signals*100)
      } else if (signals < 9000) && (signals > 1000) {
        signals = (signals)
      }

      if signals <= 1000 {
        signals = 1000
      }

      return uint32(signals)
    }
    return 0
}

func pwm2defer(wg *sync.WaitGroup, percent2 uint32) {
  var lpwm2, _ = c.RequestLine(rpi.GPIO20, gpiod.AsOutput(0))

  if os.Getenv("pwmSTOP") == "True" {
     defer wg.Done()
     os.Exit(0)
  }

  if ((percent2 >= 1) && (percent2 < 10000))  {
     time.Sleep((time.Duration(percent2) * time.Microsecond))
     // set value 1
     lpwm2.SetValue(1)
     time.Sleep((time.Duration(10) * time.Microsecond))
     // set value 0
     lpwm2.SetValue(0)

   }

     defer wg.Done()
     defer lpwm2.Close()

}

func pwm1defer(wg *sync.WaitGroup,  percent1 uint32) {
   var lpwm1, _ = c.RequestLine(rpi.GPIO16, gpiod.AsOutput(0))

   if os.Getenv("pwmSTOP") == "True" {
      defer wg.Done()
      os.Exit(0)
   }

   if ((percent1 >= 1) && (percent1 < 10000))  {
      time.Sleep((time.Duration(percent1) * time.Microsecond))
      // set value 1
      lpwm1.SetValue(1)
      time.Sleep((time.Duration(10) * time.Microsecond))
      // set value 0
      lpwm1.SetValue(0)
   }

    defer wg.Done()
    defer lpwm1.Close()


}

func eventHandler(evt gpiod.LineEvent) {

  if evt.Type == gpiod.LineEventFallingEdge {
    var wg sync.WaitGroup
    wg.Add(2)
    go pwm1defer(&wg, percent1)
    go pwm2defer(&wg, percent2)
    wg.Wait()
  }

}

func pwm(pins int, pins2 int) (string, error) {

  fmt.Println("init pwm process...")

     offset := rpi.GPIO6
     l, err := c.RequestLine(offset,
		             gpiod.WithPullUp,
	                     gpiod.WithBothEdges,
                             gpiod.WithEventHandler(eventHandler))
     if err != nil {
      fmt.Printf("RequestLine returned error: %s\n", err)
      if err == syscall.Errno(22) {
         fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
      } 
     os.Exit(1) 
  }
	
  // In a real application the main thread would do something useful here.
  // But we'll just run for a minute then exit.
  fmt.Println("please wait command usage :\n cmd : pwm[1-2] percent[0-100%]\n cmd : [reset] for disable all comand \n cmd : [stop] for quit program ")
  input := bufio.NewScanner(os.Stdin)
  for input.Scan() {

    fmt.Println(os.Getenv("pwmSTOP"))

    if input.Text() == "stop" || os.Getenv("pwmSTOP") == "True" {
	percent1 = uint32(0)
	percent2 = uint32(0)
	break
    } else {

    if strings.Contains(input.Text(), "pwm1") {
        detectpercent1 := strings.Split(input.Text(), " ")[1]
        percent1 = getInt(detectpercent1)

	if ((percent1 == 8999) || (percent1 == 9000)) {
	   percent1 = 0
	}

	fmt.Println("debug pwm1:", percent1)
      }

      if strings.Contains(input.Text(), "pwm2") {
        detectpercent2 := strings.Split(input.Text(), " ")[1]
        percent2 = getInt(detectpercent2)
        if ((percent2 == 8999) || (percent2 == 9000)) {
          percent2 = 0
        }

        fmt.Println("debug pwm2:", percent2)
      }

      if strings.Contains(input.Text(), "reset") {
        percent1 = uint32(0)
        percent2 = uint32(0)
      }
    }

  }
	
  l.Close()
  c.Close()

  return "", nil
}

// Watches GPIO 7 and reports when it changes state.
func main() {
    pwm(rpi.GPIO16, rpi.GPIO20)
}
