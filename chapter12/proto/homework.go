package proto

import (
	"io"
	"net-c12/homework/v1"

	"google.golang.org/protobuf/proto"
)

func Load(r io.Reader) ([]*homework.Chore, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var chores homework.Chores
	return chores.Chores, proto.Unmarshal(b, &chores)
}

func Flush(w io.Writer, chores []*homework.Chore) error {
	b, err := proto.Marshal(&homework.Chores{Chores: chores})
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
