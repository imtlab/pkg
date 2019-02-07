package utils

import (
	"math/rand"
	"time"
	"strconv"
)

/*
	Package utils contains handy commonly used utility functions.
*/

func ComputeProgressDivisor(itemCount int, maxIndicatorCount int) int {
	quotient := itemCount / maxIndicatorCount
	if 0 != itemCount % maxIndicatorCount {
		quotient++
	}
	return quotient
}

func ConvertTimeIsoToUnix(s string) (ts int64, err error) {
	var dt time.Time
	if dt, err = time.Parse(time.RFC3339, s); nil == err {
		ts = dt.Unix()
	}
	return
}

/*
A Duration represents the elapsed time between two instants as an int64 nanosecond count.
type Duration int64

func Since(t Time) Duration		It is shorthand for time.Now().Sub(t).
func Until(t Time) Duration		It is shorthand for t.Sub(time.Now()).


A Time represents an instant in time with nanosecond precision.
Time instants can be compared using the Before, After, and Equal methods.
The Sub method subtracts two instants, producing a Duration.
The Add method adds a Time and a Duration, producing a Time.
*/
func GetSleepDuration(lower, upper int) time.Duration {
	return time.Second * time.Duration(lower + rand.Intn(1+upper-lower))
}

func Sleep(timeStarted time.Time, durationSleep time.Duration) {
	durationSinceStarted := time.Since(timeStarted)
	if durationSinceStarted < durationSleep {
		//	adjust durationSleep to accommodate the time already spent
		durationSleep -= durationSinceStarted
		time.Sleep(durationSleep)
	}
	//	else don't sleep
}


func XStringFromXInt(xi []int) []string {
	/*	What a drag that we have create a slice of string from the slice of int
		If only there was a strings.Join(a []interface{}, sep string) string that would internally call the interface's String() method.
	*/
	//	Create a string slice using strconv.Itoa().  Itoa is shorthand for FormatInt(int64(i), 10).
	xs := make([]string, 0, len(xi))
	for _, value := range xi {
		xs = append(xs, strconv.Itoa(value))
	}
	return xs
}

/*
func getPaths() {
	//	check existence of source dir (and make sure it's a directory).
	if fileInfoSrc, err := os.Stat(sPathDirSrc); nil == err {
		if !fileInfoSrc.IsDir() {
			fmt.Printf("ERROR: \"%v\" is not a directory\n", sPathDirSrc)
			return
		}
	} else {
		fmt.Println(err)
		return
	}

	//	check existence of destination dir (and make sure it's a directory). create it if it ain't there
	if fileInfoDst, err := os.Stat(sPathDirDst); nil == err {
		if !fileInfoDst.IsDir() {
			fmt.Printf("ERROR: \"%v\" is not a directory\n", fileInfoDst)
			return
		}
	} else {
		if os.IsNotExist(err) {
			//	create the directory
			if err = os.MkdirAll(sPathDirDst, os.ModePerm); nil != err {
				fmt.Println(err)
				return
			}
		}
	}
}
*/
