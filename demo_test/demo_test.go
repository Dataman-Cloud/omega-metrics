package demo_test

/**
 *   https://github.com/smartystreets/goconvey
 */
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// run go test -v demo_test.go in command line to test the whole file
// or you can run go test -v -run TestIntegerStuff in command line to test one function
func TestIntegerStuff(t *testing.T) {
	Convey("Given some integer with a starting value", t, func() {
		x := 1

		Convey("When the integer is incremented", func() {
			x++

			Convey("The value should be greater by one", func() {
				// for more usage about So function: https://github.com/smartystreets/goconvey/wiki/Assertions
				So(x, ShouldEqual, 2)
			})
		})
	})
}

func TestString(t *testing.T) {
	Convey("Comparing two variables", t, func() {
		myVar := "Hello, world!"

		Convey(`"Asdf" should NOT equal "qwerty"`, func() {
			So("Asdf", ShouldNotEqual, "qwerty")
		})

		Convey("myVar should not be nil", func() {
			So(myVar, ShouldNotBeNil)
		})
	})
}

// user defined function
func shouldScareGophersMoreThan(actual interface{}, expected ...interface{}) string {
	if actual == "BOO!" && expected[0] == "boo" {
		return ""
	} else {
		return "Ha! You'll have to get a lot friendlier with the capslock if you want to scare a gopher!"
	}
}

func TestUserDefinedFunction(t *testing.T) {
	Convey("All caps always makes text more meaningful", t, func() {
		So("BOO!", shouldScareGophersMoreThan, "boo")
	})
}

func TestHttpRecorder(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something failed", http.StatusInternalServerError)
	}

	req, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	if err != nil {
		t.Logf(err.Error())
	}

	w := httptest.NewRecorder()

	handler(w, req)

	Convey("the response code is 500", t, func() {
		So(w.Code, ShouldEqual, 500)
	})

	Convey("the response body is 'something failed'", t, func() {
		So(w.Body.String(), ShouldEqual, "something failed")
	})
}

func TestHttpServer(t *testing.T) {
	handle := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, client")
	}
	ts := httptest.NewServer(http.HandlerFunc(handle))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		fmt.Println(err.Error())
	}

	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	Convey("receive 'hello, client' from Server", t, func() {
		So(string(text), ShouldEqual, "hello, client")
	})
}
