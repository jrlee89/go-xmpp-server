package main

import (
	"crypto/rand"
	"encoding/binary"
	"log"
)

func rng() uint64 {
	var buf [8]byte
	if _, err := rand.Reader.Read(buf[:]); err != nil {
		log.Panic("Failed to read random bytes: " + err.Error())
	}
	return binary.LittleEndian.Uint64(buf[:])
}
