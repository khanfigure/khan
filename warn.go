package duck

import (
	"fmt"
)

func Warnf(str string, args ...interface{}) {
	fmt.Printf("[WARN] "+str+"\n", args...)
}
