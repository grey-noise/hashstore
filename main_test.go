package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

var out io.Writer = os.Stdout

func TestListRuns(t *testing.T) {

	buf := &bytes.Buffer{}
	out = buf
	dbname = "hashstore.db"
	listRuns(nil)
	fmt.Printf("ici %s", buf.String())

	/*
		tests := []struct {
			name   string
			fields Bloomsky
			want   string
		}{
			{"Test1", mybloomskyTest1, "Thuin"},
			{"Test2", mybloomskyTest2, "Paris"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.fields.listRuns(nil); got != tt.want {
					t.Errorf("BloomskyStructure.GetCity() = %v, want %v", got, tt.want)
				}
			})
		}*/
}
