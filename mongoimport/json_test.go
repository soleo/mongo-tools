package mongoimport

import (
	"bytes"
	"github.com/mongodb/mongo-tools/common/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"testing"
)

func TestJSONArrayStreamDocument(t *testing.T) {
	testutil.VerifyTestType(t, testutil.UnitTestType)
	Convey("With a JSON array input reader", t, func() {
		var jsonFile, fileHandle *os.File
		Convey("an error should be thrown if a plain JSON document is supplied", func() {
			contents := `{"a": "ae"}`
			r := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
			So(r.StreamDocument(true, make(chan bson.D, 1)), ShouldNotBeNil)
		})

		Convey("reading a JSON object that has no opening bracket should "+
			"error out", func() {
			contents := `{"a":3},{"b":4}]`
			r := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
			So(r.StreamDocument(true, make(chan bson.D, 1)), ShouldNotBeNil)
		})

		Convey("JSON arrays that do not end with a closing bracket should "+
			"error out", func() {
			contents := `[{"a": "ae"}`
			r := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
			docChan := make(chan bson.D, 1)
			So(r.StreamDocument(true, docChan), ShouldNotBeNil)
			// though first read should be fine
			So(<-docChan, ShouldResemble, bson.D{bson.DocElem{"a", "ae"}})
		})

		Convey("an error should be thrown if a plain JSON file is supplied", func() {
			fileHandle, err := os.Open("testdata/test_plain.json")
			So(err, ShouldBeNil)
			r := NewJSONInputReader(true, fileHandle, 1)
			So(r.StreamDocument(true, make(chan bson.D, 50)), ShouldNotBeNil)
		})

		Convey("array JSON input file sources should be parsed correctly and "+
			"subsequent imports should parse correctly", func() {
			// TODO: currently parses JSON as floats and not ints
			expectedReadOne := bson.D{
				bson.DocElem{"a", 1.2},
				bson.DocElem{"b", "a"},
				bson.DocElem{"c", 0.4},
			}
			expectedReadTwo := bson.D{
				bson.DocElem{"a", 2.4},
				bson.DocElem{"b", "string"},
				bson.DocElem{"c", 52.9},
			}
			fileHandle, err := os.Open("testdata/test_array.json")
			So(err, ShouldBeNil)
			r := NewJSONInputReader(true, fileHandle, 1)
			docChan := make(chan bson.D, 50)
			So(r.StreamDocument(true, docChan), ShouldBeNil)
			So(<-docChan, ShouldResemble, expectedReadOne)
			So(<-docChan, ShouldResemble, expectedReadTwo)
		})

		Reset(func() {
			jsonFile.Close()
			fileHandle.Close()
		})
	})
}

func TestJSONPlainStreamDocument(t *testing.T) {
	testutil.VerifyTestType(t, testutil.UnitTestType)
	Convey("With a plain JSON input reader", t, func() {
		var jsonFile, fileHandle *os.File
		Convey("string valued JSON documents should be imported properly", func() {
			contents := `{"a": "ae"}`
			expectedRead := bson.D{bson.DocElem{"a", "ae"}}
			r := NewJSONInputReader(false, bytes.NewReader([]byte(contents)), 1)
			docChan := make(chan bson.D, 1)
			So(r.StreamDocument(true, docChan), ShouldBeNil)
			So(<-docChan, ShouldResemble, expectedRead)
		})

		Convey("several string valued JSON documents should be imported "+
			"properly", func() {
			contents := `{"a": "ae"}{"b": "dc"}`
			expectedReadOne := bson.D{bson.DocElem{"a", "ae"}}
			expectedReadTwo := bson.D{bson.DocElem{"b", "dc"}}
			r := NewJSONInputReader(false, bytes.NewReader([]byte(contents)), 1)
			docChan := make(chan bson.D, 2)
			So(r.StreamDocument(true, docChan), ShouldBeNil)
			So(<-docChan, ShouldResemble, expectedReadOne)
			So(<-docChan, ShouldResemble, expectedReadTwo)
		})

		Convey("number valued JSON documents should be imported properly", func() {
			contents := `{"a": "ae", "b": 2.0}`
			expectedRead := bson.D{bson.DocElem{"a", "ae"}, bson.DocElem{"b", 2.0}}
			r := NewJSONInputReader(false, bytes.NewReader([]byte(contents)), 1)
			docChan := make(chan bson.D, 1)
			So(r.StreamDocument(true, docChan), ShouldBeNil)
			So(<-docChan, ShouldResemble, expectedRead)
		})

		Convey("JSON arrays should return an error", func() {
			contents := `[{"a": "ae", "b": 2.0}]`
			r := NewJSONInputReader(false, bytes.NewReader([]byte(contents)), 1)
			So(r.StreamDocument(true, make(chan bson.D, 50)), ShouldNotBeNil)
		})

		Convey("plain JSON input file sources should be parsed correctly and "+
			"subsequent imports should parse correctly", func() {
			expectedReads := []bson.D{
				bson.D{
					bson.DocElem{"a", 4},
					bson.DocElem{"b", "string value"},
					bson.DocElem{"c", 1},
				},
				bson.D{
					bson.DocElem{"a", 5},
					bson.DocElem{"b", "string value"},
					bson.DocElem{"c", 2},
				},
				bson.D{
					bson.DocElem{"a", 6},
					bson.DocElem{"b", "string value"},
					bson.DocElem{"c", 3},
				},
			}
			fileHandle, err := os.Open("testdata/test_plain.json")
			So(err, ShouldBeNil)
			r := NewJSONInputReader(false, fileHandle, 1)
			docChan := make(chan bson.D, len(expectedReads))
			So(r.StreamDocument(true, docChan), ShouldBeNil)
			for i := 0; i < len(expectedReads); i++ {
				for j, readDocument := range <-docChan {
					So(readDocument.Name, ShouldEqual, expectedReads[i][j].Name)
					So(readDocument.Value, ShouldEqual, expectedReads[i][j].Value)
				}
			}
		})

		Reset(func() {
			jsonFile.Close()
			fileHandle.Close()
		})
	})
}

func TestReadJSONArraySeparator(t *testing.T) {
	testutil.VerifyTestType(t, testutil.UnitTestType)
	Convey("With an array JSON input reader", t, func() {
		Convey("reading a JSON array separator should consume [",
			func() {
				contents := `[{"a": "ae"}`
				jsonImporter := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(jsonImporter.readJSONArraySeparator(), ShouldBeNil)
				// at this point it should have consumed all bytes up to `{`
				So(jsonImporter.readJSONArraySeparator(), ShouldNotBeNil)
			})
		Convey("reading a closing JSON array separator without a "+
			"corresponding opening bracket should error out ",
			func() {
				contents := `]`
				jsonImporter := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(jsonImporter.readJSONArraySeparator(), ShouldNotBeNil)
			})
		Convey("reading an opening JSON array separator without a "+
			"corresponding closing bracket should error out ",
			func() {
				contents := `[`
				jsonImporter := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(jsonImporter.readJSONArraySeparator(), ShouldBeNil)
				So(jsonImporter.readJSONArraySeparator(), ShouldNotBeNil)
			})
		Convey("reading an opening JSON array separator with an ending "+
			"closing bracket should return EOF",
			func() {
				contents := `[]`
				jsonImporter := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(jsonImporter.readJSONArraySeparator(), ShouldBeNil)
				So(jsonImporter.readJSONArraySeparator(), ShouldEqual, io.EOF)
			})
		Convey("reading an opening JSON array separator, an ending closing "+
			"bracket but then additional characters after that, should error",
			func() {
				contents := `[]a`
				jsonImporter := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(jsonImporter.readJSONArraySeparator(), ShouldBeNil)
				So(jsonImporter.readJSONArraySeparator(), ShouldNotBeNil)
			})
		Convey("reading invalid JSON objects between valid objects should "+
			"error out",
			func() {
				contents := `[{"a":3}x{"b":4}]`
				r := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				docChan := make(chan bson.D, 1)
				So(r.StreamDocument(true, docChan), ShouldNotBeNil)
				// read first valid document
				<-docChan
				So(r.readJSONArraySeparator(), ShouldNotBeNil)
			})
		Convey("reading invalid JSON objects after valid objects but between "+
			"valid objects should error out",
			func() {
				contents := `[{"a":3},b{"b":4}]`
				r := NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(r.StreamDocument(true, make(chan bson.D, 1)), ShouldNotBeNil)
				contents = `[{"a":3},,{"b":4}]`
				r = NewJSONInputReader(true, bytes.NewReader([]byte(contents)), 1)
				So(r.StreamDocument(true, make(chan bson.D, 1)), ShouldNotBeNil)
			})
	})
}

func TestJSONConvert(t *testing.T) {
	testutil.VerifyTestType(t, testutil.UnitTestType)
	Convey("With a JSON input reader", t, func() {
		Convey("calling convert on a JSONConverter should return the expected BSON document", func() {
			jsonConverter := JSONConverter{
				data:  []byte(`{field1:"a",field2:"b",field3:"c"}`),
				index: uint64(0),
			}
			expectedDocument := bson.D{
				bson.DocElem{"field1", "a"},
				bson.DocElem{"field2", "b"},
				bson.DocElem{"field3", "c"},
			}
			document, err := jsonConverter.Convert()
			So(err, ShouldBeNil)
			So(document, ShouldResemble, expectedDocument)
		})
	})
}
