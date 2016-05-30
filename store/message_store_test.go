package store

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func Test_Fetch(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "message_store_test")
	//defer os.RemoveAll(dir)

	// when i store a message
	store := NewFileMessageStore(dir)
	a.NoError(store.Store("p1", uint64(1), []byte("aaaaaaaaaa")))
	a.NoError(store.Store("p1", uint64(2), []byte("bbbbbbbbbb")))
	a.NoError(store.Store("p2", uint64(1), []byte("1111111111")))
	a.NoError(store.Store("p2", uint64(2), []byte("2222222222")))

	testCases := []struct {
		description     string
		req             FetchRequest
		expectedResults []string
	}{
		{`match in partition 1`,
			FetchRequest{Partition: "p1", StartId: 2, Count: 1},
			[]string{"bbbbbbbbbb"},
		},
		{`match in partition 2`,
			FetchRequest{Partition: "p2", StartId: 2, Count: 1},
			[]string{"2222222222"},
		},
	}

	for _, testcase := range testCases {
		testcase.req.MessageC = make(chan MessageAndId)
		testcase.req.ErrorCallback = make(chan error)
		testcase.req.StartCallback = make(chan int)

		messages := []string{}

		store.Fetch(testcase.req)

		select {
		case numberOfResults := <-testcase.req.StartCallback:
			a.Equal(len(testcase.expectedResults), numberOfResults)
		case <-time.After(time.Second):
			a.Fail("timeout")
			return
		}

	loop:
		for {
			select {
			case msg, open := <-testcase.req.MessageC:
				if !open {
					break loop
				}
				messages = append(messages, string(msg.Message))
			case err := <-testcase.req.ErrorCallback:
				a.Fail(err.Error())
				break loop
			case <-time.After(time.Second):
				a.Fail("timeout")
				return
			}
		}
		a.Equal(testcase.expectedResults, messages, "Tescase: "+testcase.description)
	}
}

func Test_MessageStore_Close(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "message_store_test")
	//defer os.RemoveAll(dir)

	// when i store a message
	store := NewFileMessageStore(dir)
	a.NoError(store.Store("p1", uint64(1), []byte("aaaaaaaaaa")))
	a.NoError(store.Store("p2", uint64(1), []byte("1111111111")))

	a.Equal(2, len(store.partitions))

	a.NoError(store.Stop())

	a.Equal(0, len(store.partitions))
}

func Test_MaxMessageId(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "message_store_test")
	//defer os.RemoveAll(dir)
	expectedMaxId := 2

	// when i store a message
	store := NewFileMessageStore(dir)
	a.NoError(store.Store("p1", uint64(1), []byte("aaaaaaaaaa")))
	a.NoError(store.Store("p1", uint64(expectedMaxId), []byte("bbbbbbbbbb")))

	maxID, err := store.MaxMessageId("p1")
	a.Nil(err, "No error should be received for partition p1")
	a.Equal(maxID, uint64(expectedMaxId), fmt.Sprintf("MaxId should be [%d]", expectedMaxId))
}

func Test_MessagePartitionReturningError(t *testing.T) {
	a := assert.New(t)

	store := NewFileMessageStore("/TestDir")
	_, err := store.partitionStore("p1")
	a.NotNil(err)
	fmt.Println(err)

	store2 := NewFileMessageStore("/")
	_, err2 := store2.partitionStore("p1")
	fmt.Println(err2)
}

func Test_FetchWithError(t *testing.T) {
	a := assert.New(t)
	store := NewFileMessageStore("/TestDir")

	chanCallBack := make(chan error, 1)
	aFetchRequest := FetchRequest{Partition: "p1", StartId: 2, Count: 1, ErrorCallback: chanCallBack}
	store.Fetch(aFetchRequest)
	err := <-aFetchRequest.ErrorCallback
	a.NotNil(err)
}

func Test_StoreWithError(t *testing.T) {
	a := assert.New(t)
	store := NewFileMessageStore("/TestDir")

	err := store.Store("p1", uint64(1), []byte("124151qfas"))
	a.NotNil(err)
}
