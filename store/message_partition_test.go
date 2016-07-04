package store

import (

	//"github.com/smancke/guble/testutil"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/smancke/guble/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_MessagePartition_scanFiles(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	a.NoError(ioutil.WriteFile(path.Join(dir, "myMessages-00000000000000420000.idx"), []byte{}, 0777))
	a.NoError(ioutil.WriteFile(path.Join(dir, "myMessages-00000000000000000000.idx"), []byte{}, 0777))
	a.NoError(ioutil.WriteFile(path.Join(dir, "myMessages-00000000000000010000.idx"), []byte{}, 0777))

	err := store.readIdxFiles()
	a.NoError(err)
	//a.Equal([]uint64{
	//	0,
	//	10000,
	//	420000,
	//}, fileIds)
}

func Test_MessagePartition_correctIdAfterRestart(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	a.NoError(store.Store(uint64(1), []byte("aaaaaaaaaa")))
	a.NoError(store.Store(uint64(2), []byte("aaaaaaaaaa")))
	a.Equal(uint64(2), fne(store.MaxMessageId()))
	a.NoError(store.Close())

	newStore, err := NewMessagePartition(dir, "myMessages")
	a.NoError(err)
	a.Equal(uint64(2), fne(newStore.MaxMessageId()))
}

func TestCreateNextAppendFiles(t *testing.T) {
	a := assert.New(t)

	// given: a store
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	// when i create append files
	a.NoError(store.createNextAppendFiles())

	//a.Equal(uint64(420000), store.appendFirstId)
	//a.Equal(uint64(429999), store.appendLastId)
	a.Equal(uint64(0x9), store.appendFileWritePosition)

	// and close the store
	a.NoError(store.Close())

	// then both files are empty
	msgs, errRead := ioutil.ReadFile(dir + "/myMessages-00000000000000420000.msg")
	a.NoError(errRead)
	a.Equal(9, len(msgs))
	a.Equal(MAGIC_NUMBER, msgs[0:8])
	a.Equal(byte(1), msgs[8])

	idx, errIdx := ioutil.ReadFile(dir + "/myMessages-00000000000000420000.idx")
	a.NoError(errIdx)
	a.Equal([]byte{}, idx)
}

func Test_Storing_Two_Messages_With_Append(t *testing.T) {
	a := assert.New(t)

	// given: a store
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)

	// when i store a message
	store, _ := NewMessagePartition(dir, "myMessages")
	a.NoError(store.Store(0x1, []byte("abc")))
	a.NoError(store.Close())

	// and I add another message with a new store
	store, _ = NewMessagePartition(dir, "myMessages")
	a.NoError(store.Store(0x2, []byte("defgh")))
	a.NoError(store.Close())

	// then both files as expected
	msgs, errRead := ioutil.ReadFile(dir + "/myMessages-00000000000000000000.msg")
	a.NoError(errRead)
	a.Equal(MAGIC_NUMBER, msgs[0:8])
	a.Equal(byte(1), msgs[8])
	a.Equal([]byte{
		3, 0, 0, 0, // len(abc) == 3
		1, 0, 0, 0, 0, 0, 0, 0, // id == 1
		'a', 'b', 'c',

		5, 0, 0, 0, // len(defgh) == 5
		2, 0, 0, 0, 0, 0, 0, 0, // id == 3
		'd', 'e', 'f', 'g', 'h', // defgh
	}, msgs[9:])

	idx, errIdx := ioutil.ReadFile(dir + "/myMessages-00000000000000000000.idx")
	a.NoError(errIdx)
	a.Equal([]byte{
		0, 0, 0, 0, 0, 0, 0, 0, // id 0: not set
		0, 0, 0, 0, // size == 0
		21, 0, 0, 0, 0, 0, 0, 0, // id 1: offset
		3, 0, 0, 0, // size == 0
		36, 0, 0, 0, 0, 0, 0, 0, // id 2: offset
		5, 0, 0, 0, // size == 5
	}, idx)
}

func Benchmark_Storing_HelloWorld_Messages(b *testing.B) {
	a := assert.New(b)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		a.NoError(store.Store(uint64(i), []byte("Hello World")))
	}
	a.NoError(store.Close())
	b.StopTimer()
}

func Benchmark_Storing_1Kb_Messages(b *testing.B) {
	a := assert.New(b)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	message := make([]byte, 1024)
	for i := range message {
		message[i] = 'a'
	}

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		a.NoError(store.Store(uint64(i), message))
	}
	a.NoError(store.Close())
	b.StopTimer()
}

func Benchmark_Storing_1MB_Messages(b *testing.B) {
	a := assert.New(b)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)
	store, _ := NewMessagePartition(dir, "myMessages")

	message := make([]byte, 1024*1024)
	for i := range message {
		message[i] = 'a'
	}

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		a.NoError(store.Store(uint64(i), message))
	}
	a.NoError(store.Close())
	b.StopTimer()
}

//func TestFirstMessageIdForFile(t *testing.T) {
//	a := assert.New(t)
//	store := &MessagePartition{}
//	a.Equal(uint64(0), store.firstMessageIdForFile(0))
//	a.Equal(uint64(0), store.firstMessageIdForFile(1))
//	a.Equal(uint64(0), store.firstMessageIdForFile(42))
//	a.Equal(uint64(7680000), store.firstMessageIdForFile(7682334))
//}

func Test_calculateFetchList(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	// allow five messages per file
	MESSAGES_PER_FILE = uint64(5)

	msgData := []byte("aaaaaaaaaa") // 10 bytes message

	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)

	store, _ := NewMessagePartition(dir, "myMessages")

	// File header: MAGIC_NUMBER + FILE_NUMBER_VERSION = 9 bytes in the file
	// For each stored message there is a 12 bytes write that contains the msgID and size

	a.NoError(store.Store(uint64(3), msgData)) // stored offset 21, size: 10
	a.NoError(store.Store(uint64(4), msgData)) // stored offset 21+10+12=43

	a.NoError(store.Store(uint64(10), msgData)) // stored offset 43+22=65

	a.NoError(store.Store(uint64(9), msgData)) // stored offset 65+22=87
	a.NoError(store.Store(uint64(5), msgData)) // stored offset 87+22=109

	// here second file will start
	a.NoError(store.Store(uint64(8), msgData))  // stored offset 21
	a.NoError(store.Store(uint64(15), msgData)) // stored offset 43
	a.NoError(store.Store(uint64(13), msgData)) // stored offset 65

	a.NoError(store.Store(uint64(22), msgData)) // stored offset 87
	a.NoError(store.Store(uint64(23), msgData)) // stored offset 109

	// third file
	a.NoError(store.Store(uint64(24), msgData)) // stored offset 21
	a.NoError(store.Store(uint64(26), msgData)) // stored offset 43

	a.NoError(store.Store(uint64(30), msgData)) // stored offset 65

	defer a.NoError(store.Close())

	testCases := []struct {
		description     string
		req             FetchRequest
		expectedResults SortedIndexList
	}{
		{`direct match`,
			FetchRequest{StartID: 3, Direction: 0, Count: 1},
			SortedIndexList{
				{3, uint64(21), 10, 0}, // messageId, offset, size, fileId
			},
		},
		{`direct match in second file`,
			FetchRequest{StartID: 8, Direction: 0, Count: 1},
			SortedIndexList{
				{8, uint64(21), 10, 1}, // messageId, offset, size, fileId,
			},
		},
		{`direct match in second file, not first position`,
			FetchRequest{StartID: 13, Direction: 0, Count: 1},
			SortedIndexList{
				{13, uint64(65), 10, 1}, // messageId, offset, size, fileId,
			},
		},
		// TODO
		// {`next entry matches`,
		// 	FetchRequest{StartID: 1, Direction: 0, Count: 1},
		// 	SortedIndexList{
		// 		{3, uint64(21), 10, 0}, // messageId, offset, size, fileId
		// 	},
		// },
		{`entry before matches`,
			FetchRequest{StartID: 5, Direction: -1, Count: 2},
			SortedIndexList{
				{4, uint64(43), 10, 0},  // messageId, offset, size, fileId
				{5, uint64(109), 10, 0}, // messageId, offset, size, fileId
			},
		},
		{`backward, no match`,
			FetchRequest{StartID: 1, Direction: -1, Count: 1},
			SortedIndexList{},
		},
		{`forward, no match (out of files)`,
			FetchRequest{StartID: 99999999999, Direction: 1, Count: 1},
			SortedIndexList{},
		},
		{`forward, no match (after last id in last file)`,
			FetchRequest{StartID: 31, Direction: 1, Count: 1},
			SortedIndexList{},
		},
		{`forward, overlapping files`,
			FetchRequest{StartID: 9, Direction: 1, Count: 3},
			SortedIndexList{
				{9, uint64(87), 10, 0},  // messageId, offset, size, fileId
				{10, uint64(65), 10, 0}, // messageId, offset, size, fileId
				{13, uint64(65), 10, 1}, // messageId, offset, size, fileId
			},
		},
		{`backward, overlapping files`,
			FetchRequest{StartID: 26, Direction: -1, Count: 4},
			SortedIndexList{
				// {15, uint64(43), 10, 1},  // messageId, offset, size, fileId
				{22, uint64(87), 10, 1},  // messageId, offset, size, fileId
				{23, uint64(109), 10, 1}, // messageId, offset, size, fileId
				{24, uint64(21), 10, 2},  // messageId, offset, size, fileId
				{26, uint64(43), 10, 2},  // messageId, offset, size, fileId
			},
		},
		{`forward, over more then 2 files`,
			FetchRequest{StartID: 5, Direction: 1, Count: 10},
			SortedIndexList{
				{5, uint64(109), 10, 0},  // messageId, offset, size, fileId
				{8, uint64(21), 10, 1},   // messageId, offset, size, fileId
				{9, uint64(87), 10, 0},   // messageId, offset, size, fileId
				{10, uint64(65), 10, 0},  // messageId, offset, size, fileId
				{13, uint64(65), 10, 1},  // messageId, offset, size, fileId
				{15, uint64(43), 10, 1},  // messageId, offset, size, fileId
				{22, uint64(87), 10, 1},  // messageId, offset, size, fileId
				{23, uint64(109), 10, 1}, // messageId, offset, size, fileId
				{24, uint64(21), 10, 2},  // messageId, offset, size, fileId
				{26, uint64(43), 10, 2},  // messageId, offset, size, fileId
			},
		},
	}

	for _, testcase := range testCases {
		testcase.req.Partition = "myMessages"
		fetchEntries, err := store.calculateFetchListNew(&testcase.req)
		a.NoError(err, "Tescase: "+testcase.description)
		a.True(matchSortedList(t, testcase.expectedResults, *fetchEntries), "Tescase: "+testcase.description)
	}
}

func matchSortedList(t *testing.T, expected, actual SortedIndexList) bool {
	if len(expected) != len(actual) {
		assert.Equal(t, len(expected), len(actual), "Invalid length")
		return false
	}

	for i, entry := range expected {
		a := actual[i]
		assert.Equal(t, *entry, *a)
		if entry.messageId != a.messageId ||
			entry.offset != a.offset ||
			entry.size != a.size ||
			entry.fileID != a.fileID {
			return false
		}
	}

	return true
}

func Test_Partition_Fetch(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)

	// when i store a message
	store, _ := NewMessagePartition(dir, "myMessages")
	store.maxMessageId = uint64(2) // hack, for test setup
	a.NoError(store.Store(uint64(3), []byte("aaaaaaaaaa")))
	a.NoError(store.Store(uint64(4), []byte("bbbbbbbbbb")))
	store.maxMessageId = uint64(9) // hack, for test setup
	a.NoError(store.Store(uint64(10), []byte("cccccccccc")))
	store.maxMessageId = MESSAGES_PER_FILE - uint64(1) // hack, for test setup
	a.NoError(store.Store(MESSAGES_PER_FILE, []byte("1111111111")))
	store.maxMessageId = MESSAGES_PER_FILE + uint64(4) // hack, for test setup
	a.NoError(store.Store(MESSAGES_PER_FILE+uint64(5), []byte("2222222222")))
	defer a.NoError(store.Close())

	testCases := []struct {
		description     string
		req             FetchRequest
		expectedResults []string
	}{
		{`direct match`,
			FetchRequest{StartID: 3, Direction: 0, Count: 1},
			[]string{"aaaaaaaaaa"},
		},
		{`direct match in second file`,
			FetchRequest{StartID: MESSAGES_PER_FILE, Direction: 0, Count: 1},
			[]string{"1111111111"},
		},
		{`next entry matches`,
			FetchRequest{StartID: 1, Direction: 0, Count: 1},
			[]string{"aaaaaaaaaa"},
		},
		{`entry before matches`,
			FetchRequest{StartID: 5, Direction: -1, Count: 1},
			[]string{"bbbbbbbbbb"},
		},
		{`backward, no match`,
			FetchRequest{StartID: 1, Direction: -1, Count: 1},
			[]string{},
		},
		{`forward, no match (out of files)`,
			FetchRequest{StartID: 99999999999, Direction: 1, Count: 1},
			[]string{},
		},
		{`forward, no match (after last id in last file)`,
			FetchRequest{StartID: MESSAGES_PER_FILE + uint64(8), Direction: 1, Count: 1},
			[]string{},
		},
		/*
			{`forward, overlapping files`,
				FetchRequest{StartId: 9, Direction: 1, Count: 3},
				[]string{"cccccccccc", "1111111111", "2222222222"},
			},
			{`backward, overlapping files`,
				FetchRequest{StartId: MESSAGES_PER_FILE + uint64(100), Direction: -1, Count: 100},
				[]string{"2222222222", "1111111111", "cccccccccc", "bbbbbbbbbb", "aaaaaaaaaa"},
			},*/
	}
	for _, testcase := range testCases {
		testcase.req.Partition = "myMessages"
		testcase.req.MessageC = make(chan MessageAndId)
		testcase.req.ErrorCallback = make(chan error)
		testcase.req.StartCallback = make(chan int)

		messages := []string{}

		store.Fetch(&testcase.req)

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

//func TestFilenameGeneration(t *testing.T) {
//	a := assert.New(t)
//
//	store := &MessagePartition{
//		basedir: "/foo/bar/",
//		name:    "myMessages",
//	}
//
//	a.Equal("/foo/bar/myMessages-00000000000000000042.msg", store.composeMsgFilename())
//	a.Equal("/foo/bar/myMessages-00000000000000000042.idx", store.composeIndexFilename(42))
//	a.Equal("/foo/bar/myMessages-00000000000000000000.idx", store.composeIndexFilename(0))
//	a.Equal("/foo/bar/myMessages-00000000000000010000.idx", store.composeIndexFilename(MESSAGES_PER_FILE))
//}

//func Test_firstMessageIdForFile(t *testing.T) {
//	a := assert.New(t)
//	store := &MessagePartition{}
//
//	a.Equal(uint64(0), store.firstMessageIdForFile(0))
//	a.Equal(uint64(0), store.firstMessageIdForFile(42))
//	a.Equal(MESSAGES_PER_FILE, store.firstMessageIdForFile(MESSAGES_PER_FILE))
//	a.Equal(MESSAGES_PER_FILE, store.firstMessageIdForFile(MESSAGES_PER_FILE+uint64(1)))
//}

func Test_calculateMaxMessageIdFromIndex2(t *testing.T) {
	a := assert.New(t)
	defer testutil.EnableDebugForMethod()()

	f := make([]*FileCacheEntry, 0)
	f = append(f, &FileCacheEntry{
		minMsgID: 1,
		maxMsgID: 9,
	})

	f = append(f, &FileCacheEntry{
		minMsgID: 10,
		maxMsgID: 28,
	})

	f = append(f, &FileCacheEntry{
		minMsgID: 29,
		maxMsgID: 35,
	})

	f = append(f, &FileCacheEntry{
		minMsgID: 17,
		maxMsgID: 30,
	})

	req := FetchRequest{StartID: 11, EndID: 31, Direction: 0, Count: 1}

	for _, ff := range f {
		tt := ff.hasStartID(&req)
		logrus.WithField("dd", tt).Info()
	}

	req2 := FetchRequest{StartID: 11, Direction: -1, Count: 1}

	for _, ff := range f {
		tt := ff.hasStartID(&req2)
		logrus.WithField("dd", tt).Info()
	}

	a.Nil(nil)
}

func Test_calculateMaxMessageIdFromIndex(t *testing.T) {
	a := assert.New(t)
	dir, _ := ioutil.TempDir("", "guble_message_partition_test")
	defer os.RemoveAll(dir)

	// when i store a message
	store, _ := NewMessagePartition(dir, "myMessages")
	a.NoError(store.Store(uint64(1), []byte("aaaaaaaaaa")))
	a.NoError(store.Store(uint64(2), []byte("bbbbbbbbbb")))

	maxMessageId, err := store.calculateMaxMessageIdFromIndex(uint64(0))
	a.NoError(err)
	a.Equal(uint64(2), maxMessageId)
}
