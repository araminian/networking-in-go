package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	storage "net-c12/gob"
	homework "net-c12/homework"
	// storage "net-c12/json"
)

var dataFile string

func init() {
	// dataFile is a flag that specifies the data file to use
	flag.StringVar(&dataFile, "file", "housework.db", "data file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [flags] [add chore, ...|complete #]
		add add comma-separated chores
		complete complete designated chore
		Flags:
		`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func load() ([]*homework.Chore, error) {

	// check if the file exists , if not return an empty list
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return make([]*homework.Chore, 0), nil
	}

	df, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := df.Close(); err != nil {
			fmt.Printf("closing data file: %v", err)
		}
	}()
	return storage.Load(df)
}

// flushes the chores in memory to your storage for persistence.
func flush(chores []*homework.Chore) error {

	// Here, you create a new file or truncate the existing file
	df, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := df.Close(); err != nil {
			fmt.Printf("closing data file: %v", err)
		}
	}()
	return storage.Flush(df, chores)
}

// display the chores on the command line.
func list() error {
	// First, you load the list of chores from storage
	chores, err := load()
	if err != nil {
		return err
	}
	if len(chores) == 0 {
		fmt.Println("You're all caught up!")
		return nil
	}
	fmt.Println("#\t[X]\tDescription")
	for i, chore := range chores {
		c := " "
		if chore.Complete {
			c = "X"
		}
		fmt.Printf("%d\t[%s]\t%s\n", i+1, c, chore.Description)
	}
	return nil
}

// add a new chore to the list
func add(s string) error {
	chores, err := load()
	if err != nil {
		return err
	}
	/*
		You want the option to add more than one chore at a time, so you split
		the incoming chore description by commas and trim each description to remove
		any leading or trailing whitespace.
	*/
	for _, chore := range strings.Split(s, ",") {
		if desc := strings.TrimSpace(chore); desc != "" {
			chores = append(chores, &homework.Chore{
				Description: desc,
			})
		}
	}
	/*
		because you want your list of chores
		to persist between executions of the application,
		you need to store the chore state on disk.
	*/
	return flush(chores)
}

// marking chore as complete
func complete(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	chores, err := load()
	if err != nil {
		return err
	}
	if i < 1 || i > len(chores) {
		return fmt.Errorf("chore %d not found", i)
	}
	chores[i-1].Complete = true
	return flush(chores)
}

func main() {
	flag.Parse()
	var err error
	switch strings.ToLower(flag.Arg(0)) {
	case "add":
		err = add(strings.Join(flag.Args()[1:], " "))
	case "complete":
		err = complete(flag.Arg(1))
	}
	if err != nil {
		log.Fatal(err)
	}
	err = list()
	if err != nil {
		log.Fatal(err)
	}
}
