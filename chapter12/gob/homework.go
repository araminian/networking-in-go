package gob

import (
	"encoding/gob"
	"io"
	homework "net-c12/homework"
)

func Load(r io.Reader) ([]*homework.Chore, error) {
	var chores []*homework.Chore
	return chores, gob.NewDecoder(r).Decode(&chores)
}

func Flush(w io.Writer, chores []*homework.Chore) error {
	return gob.NewEncoder(w).Encode(chores)
}
