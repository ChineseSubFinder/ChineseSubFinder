package my_util

import (
	"math/rand"
	"time"
)

func RandomSecondDuration(min, max int32) time.Duration {
	tmp := random.Int31n(max-min) + min
	return time.Duration(tmp) * time.Second
}

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
)
