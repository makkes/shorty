package main

import "math/rand"

func randStringBytes(n int) []byte {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func keygen(keybuffer chan<- []byte) {
	for {
		keybuffer <- randStringBytes(10)
	}
}
