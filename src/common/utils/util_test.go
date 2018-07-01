package utils

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"
)

type TimeCheckStruct struct {
	TimeString string
	TmObj      time.Time
}

var ConfigTimeFormat = []string{
	"dd/mmm/yyyy:HH:MM:SS",
	"dd/mmm/yyyy HH:MM:SS",
	"yyyy-mm-ddTHH:MM:SS",
	"dd-mm-yyyy HH:MM:SS",
	"yyyy-mm-dd HH:MM:SS",
	"yyyy/mm/dd HH:MM:SS",
	"yyyymmdd HH:MM:SS",
	"mmm dd HH:MM:SS",
}

func (this *TimeCheckStruct) Check(configTF string) bool {
	pat, timeFormat := GetPatAndTimeFormat(configTF)
	reg := regexp.MustCompile(pat)
	t := reg.FindString(this.TimeString)
	if len(strings.TrimSpace(t)) == 0 {
		fmt.Printf("\tCannot get TimeString, [line:%s][timeFormat:%s]\n", this.TimeString, timeFormat)
		return false
	}
	tm, err := time.Parse(timeFormat, t)
	if err != nil {
		fmt.Printf("Not Equal : this[%v] error[%v]\n", this, err)
	}
	if tm.Month() == this.TmObj.Month() &&
		tm.Day() == this.TmObj.Day() &&
		tm.Hour() == this.TmObj.Hour() &&
		tm.Minute() == this.TmObj.Minute() &&
		tm.Second() == this.TmObj.Second() {
		if configTF == "mmm dd HH:MM:SS" {
			return true
		} else if tm.Year() == this.TmObj.Year() {
			return true
		}
	}
	fmt.Printf("Not Equal : tm[%v] Checkstruct[%v]\n", tm, this)
	return false
}

func initTestBody(timeFormat string) []*TimeCheckStruct {
	var retList []*TimeCheckStruct
	tmsNow := time.Now().Unix() + 50*365*24*3600
	tmStart, err := time.Parse("2006/01/02 15:04:05", "2000/01/02 01:02:03")
	if err != nil {
		log.Fatal("parse tmsStart failed : %v", err)
	}
	tmsStart := tmStart.Unix()
	//遍历每天的时间戳
	tms := tmsStart
	for ; tms < tmsNow; tms = tms + 24*3600 {
		tm := time.Unix(tms, 0)
		tmsString := tm.Format(timeFormat)
		var tmp TimeCheckStruct
		tmp.TimeString = tmsString
		tmp.TmObj = tm
		retList = append(retList, &tmp)
	}
	return retList
}

func TestTimeFormat(t *testing.T) {
	for _, ctf := range ConfigTimeFormat {
		fmt.Printf("\nNow test config format : %s\n", ctf)
		_, tf := GetPatAndTimeFormat(ctf)
		checkBody := initTestBody(tf)
		fmt.Println("\tInit test body Done")
		for _, body := range checkBody {
			if !body.Check(ctf) {
				fmt.Printf("\tFailed : %v\n", body)
				panic("Test Failed!")
			}
		}
		fmt.Println("\tTest Success!")
	}
}
