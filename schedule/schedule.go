package schedule

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mlctrez/goapp-sprinkler/beagleio"
	"github.com/nats-io/go-nats"
	"github.com/robfig/cron/v3"
)

type Schedule struct {
	chron                *cron.Cron
	beagle               *beagleio.Api
	natsConn             *nats.Conn
	lscSubscription      *nats.Subscription
	sprinklerRunningChan chan struct{}
}

func New() (*Schedule, error) {

	natsConn, err := nats.Connect("hats://goservice:19201")
	if err != nil {
		return nil, err
	}
	s := &Schedule{chron: cron.New(), beagle: beagleio.New(), natsConn: natsConn}
	err = s.subscribeLightStateChange()
	if err != nil {
		s.natsConn.Close()
		return nil, err
	}
	_, err = s.chron.AddFunc("0 1 * * *", s.SprinklerCheckDate)
	return s, err
}

const DateOnly = "20060102"
const LastDateFile = "last.date"

func (s *Schedule) SprinklerCheckDate() {
	var lastRun time.Time
	if file, err := os.ReadFile(LastDateFile); os.IsNotExist(err) {
		lastRun = time.Now().Add(-48 * time.Hour)
	} else {
		lastRun, err = time.Parse(DateOnly, strings.TrimSpace(string(file)))
		if err != nil {
			fmt.Println("error reading last date file, exiting schedule")
			return
		}
	}
	if time.Since(lastRun) >= 48*time.Hour {
		err := os.WriteFile(LastDateFile, []byte(time.Now().Format(DateOnly)), 0644)
		if err != nil {
			fmt.Println("error writing last date file, exiting schedule")
			return
		}
		s.SprinklerRun()
	}
}

var DurationMap = map[string]time.Duration{
	"30": time.Minute * 30,
	"15": time.Minute * 15,
}

var DurationMapDev = map[string]time.Duration{
	"30": time.Second * 3,
	"15": time.Second * 3,
}

var pinsAndTimes = []string{"0:30", "1:30", "2:30", "3:30", "4:30", "5:15"}

func logStartStop() func() {
	start := time.Now()
	return func() {
		stop := time.Now()
		fmt.Printf("start %s stop %s duration %4.2f minutes\n",
			start.Format(time.Kitchen), stop.Format(time.Kitchen),
			stop.Sub(start).Minutes())
	}
}

func (s *Schedule) SprinklerRun() {

	if s.sprinklerRunningChan != nil {
		return
	}

	defer logStartStop()()

	s.beagle.PinsOff()
	defer s.beagle.PinsOff()

	s.sprinklerRunningChan = make(chan struct{}, 2)
	defer func() {
		close(s.sprinklerRunningChan)
		s.sprinklerRunningChan = nil
	}()

	for _, pat := range pinsAndTimes {
		pats := strings.Split(pat, ":")

		pin := pats[0]
		duration := DurationMap[pats[1]]

		if os.Getenv("DEV") != "" {
			duration = DurationMapDev[pats[1]]
		}

		fmt.Printf("turning on pin %s for %v\n", pin, duration)
		timer := time.NewTimer(duration)

		s.beagle.ChangePin(pin, "on")

		select {
		case <-timer.C:
			s.beagle.ChangePin(pin, "off")
			if os.Getenv("DEV") == "" {
				time.Sleep(30 * time.Second)
			}
		case <-s.sprinklerRunningChan:
			fmt.Println("echo turned off the sprinkler")
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}

}

func (s *Schedule) Stop() {
	if s.lscSubscription != nil {
		err := s.lscSubscription.Unsubscribe()
		if err != nil {
			fmt.Println(err)
		}
	}
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	s.chron.Stop()
}

func (s *Schedule) subscribeLightStateChange() error {
	if subscribe, err := s.natsConn.Subscribe("lightStateChange", s.lightStateChange); err != nil {
		return err
	} else {
		s.lscSubscription = subscribe
		return nil
	}
}

func (s *Schedule) lightStateChange(msg *nats.Msg) {
	lsc := &LightStateChange{}
	err := json.Unmarshal(msg.Data, lsc)
	if err != nil {
		log.Println(err)
		return
	}

	if lsc.LightID == "aa2f65ccc32c03efc1d4d91a86ee03414f3b7893f6ce1b7e0020088122d0df61" {
		if lsc.StateRequest.On {
			if s.sprinklerRunningChan == nil {
				fmt.Println("echo turned on the sprinkler")
				go s.SprinklerRun()
			}
			return
		}
		if s.sprinklerRunningChan != nil {
			s.sprinklerRunningChan <- struct{}{}
		}
	}

}

type LightStateChange struct {
	GroupID      string       `json:"groupID"`
	LightID      string       `json:"lightID"`
	StateRequest StateRequest `json:"stateRequest"`
}

type StateRequest struct {
	On  bool  `json:"on"`
	Bri int32 `json:"bri"`
}
