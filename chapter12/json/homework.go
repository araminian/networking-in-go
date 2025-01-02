package json

import (
	"encoding/json"
	"io"
	homework "net-c12/homework"
)

/*
The Load function passes the io.Reader to the json.NewDecoder function
and returns a decoder. You then call the decoder’s Decode method, passing
it a pointer to the chores slice. The decoder reads JSON from the io.Reader,
deserializes it, and populates the chores slice.
*/
func Load(r io.Reader) ([]*homework.Chore, error) {
	var chores []*homework.Chore
	return chores, json.NewDecoder(r).Decode(&chores)
}

/*
The Flush function accepts an io.Writer and a chores slice. It then passes
the io.Writer to the json.NewEncoder function, which returns an encoder.
You pass the chores slice to the encoder’s Encode function, which serializes
the chores slice into JSON and writes it to the io.Writer.
*/
func Flush(w io.Writer, chores []*homework.Chore) error {
	return json.NewEncoder(w).Encode(chores)
}
