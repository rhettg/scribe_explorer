package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bufio"
	"http"
	"websocket"
	"log"
	"json"

)

type LineReader interface {
	ReadLine() (line []byte, isPrefix bool, err os.Error)
}

var (
	rangerDataChan chan (chan JSONData)
	allChannels [](chan JSONData)
)

func init() {
	rangerDataChan = make(chan (chan JSONData))
}

func acceptChannels() {
	for {
		newChannel := <- rangerDataChan
		log.Printf("Adding new channel to data stream")
		allChannels = append(allChannels, newChannel)
	}
}

func streamData(lineStream LineReader) {
	for {
		line, isPrefix, err := lineStream.ReadLine()
		if err != nil {
			if err == os.EOF {
				break
			}
			log.Printf("Failed on line stream", err)
			break
		}
		if isPrefix {
			log.Fatal("PREFIX!!")
		}

		// We have fairly reliable looking chunk of data, try to decode it
		var data JSONData
		err = json.Unmarshal(line, &data)
		if err != nil {
			log.Printf("Failure to decode: %s", err)
			log.Println(string(line))
			log.Println()
			continue
		}

		// Now deliver this fine chunk of ranger data to each of our listeners
		for _, webChannel := range allChannels {
			webChannel <- data
		}
	}
	log.Printf("All done with data stream")
}

func CreateTestDataStream(fileName string) io.Reader {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Failed to open file", err)
	}
	// defer file.Close()

	return file
}

func ServeWS(socket *websocket.Conn) {
	// Create a new channel to receive data on
	dataChan := make(chan JSONData)
	rangerDataChan <- dataChan
	
	for {
		data := <-dataChan
		
		uniqueRequestID, ok := GetDeep("unique_request_id", data)

		if !ok {
			log.Print("Missing unique request id")
		}

		dirtySession := "NA"
		dirtySessionBool, ok := GetDeep("extra.dirty_session", data)
		if !ok {
			log.Print("Missing dirty session")	
		} else {
			dirtySession = fmt.Sprintf("%t", dirtySessionBool)
		}

		socket.Write([]byte(fmt.Sprintf("%s: %s", uniqueRequestID, dirtySession)))
	}
}

func ServePage(writer http.ResponseWriter, request *http.Request) {
	//file, err := os.Open("default.do");
	log.Println("Starting /")
	
	//defer file.Close();
	contents, err := ioutil.ReadFile("index.html")
	if err != nil {
		fmt.Printf("Error opening file");
		return
	}

	writer.Write(contents);

}


func main() {
	log.Println("Starting up")
	
	go acceptChannels()

	stream := CreateStream("scribe-stagea.local.yelpcorp.com:3535")
	//stream := CreateTestDataStream("test_data/ranger_sample.json")
	lineStream, err := bufio.NewReaderSize(stream, 1024*32)
	if err != nil {
		log.Fatal("Failed to create reader", err)
	}

	go streamData(lineStream)
	
	http.Handle("/", http.HandlerFunc(ServePage));
	http.Handle("/ws", websocket.Handler(ServeWS));

	err = http.ListenAndServe(":8080", nil);
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}