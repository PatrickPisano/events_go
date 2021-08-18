package rand

import (
"math/rand"
"time"
)

func RandString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	rand.Seed(time.Now().UnixNano()) // todo:: find out if to do this here, may be in the init func

	bb := make([]byte, n)
	for b := range bb {
		bb[b] = letters[rand.Intn(len(letters))]
	}
	return string(bb)
}

