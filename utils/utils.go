package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

var Logger = log.New(os.Stdout, "http: ", log.LstdFlags)
var logger = Logger

func BuildCFlink(region, stackID string) string {
	return "https://" + region + ".console.aws.amazon.com/cloudformation/home?region=" + region + "#/stacks/stackinfo?stackId=" + stackID
}
func ShortMiniteUniqueID() string {
	timeStr := time.Now().UTC().Format("200601021504")
	timeInt, _ := strconv.ParseInt(timeStr, 10, 64)
	timeHex := strconv.FormatInt(int64(timeInt), 36)
	return timeHex
}
