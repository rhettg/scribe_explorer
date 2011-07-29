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
	"net"
	"net/textproto"
)

type LineReader interface {
	ReadLine() (line []byte, isPrefix bool, err os.Error)
}

var (
	rangerDataChan chan (chan JSONData)
	allChannels    [](chan JSONData)
)

func init() {
	rangerDataChan = make(chan (chan JSONData))
}

func acceptChannels() {
	for {
		newChannel := <-rangerDataChan
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
	jsonStream := NewJSONConn(socket)

	ServeStream(jsonStream)
}

func ServeStream(stream *JSONConn) {
	// Create a new channel to receive data on
	dataChan := make(chan JSONData)
	rangerDataChan <- dataChan

	// Get our query from the client
	query, err := stream.ReadJSON()
	if err != nil {
		log.Fatal("Failed to read from client", err)
	}

	displayFields := []string{}
	for _, fieldValue := range query.(map[string] interface{})["fields"].([]interface{}) {
		log.Printf("Field: ", fieldValue)
		displayFields = append(displayFields, fieldValue.(string))
	}

	filters := []Filter{}
	for _, statement := range query.(map[string] interface{})["filters"].([]interface{}) {
		log.Printf("Statement: ", statement)
		filter, ok := ParseStatement(statement.(string))
		if ok {
			log.Printf("Parsed to filter: %v", filter)
			filters = append(filters, filter)
		}
	}


	for {
		data := <-dataChan
		
		if !PassesAllFilters(data, filters) {
			continue
		}
		outputMap := map[string] interface{}{}
		
		for _, fieldValue := range displayFields {
			outputMap[fieldValue], _ = GetDeep(fieldValue, data)
		}

		err := stream.WriteJSON(outputMap)
		if err != nil {
			log.Fatal("Failed to write", err)
		}
	}
}

func ServePage(writer http.ResponseWriter, request *http.Request) {
	//file, err := os.Open("default.do");
	log.Println("Starting /")

	//defer file.Close();
	contents, err := ioutil.ReadFile("index.html")
	if err != nil {
		fmt.Printf("Error opening file")
		return
	}

	writer.Write(contents)

}

func serveTCP(conn *textproto.Conn) {
	defer conn.Close()
	
	for {
		cmd, err := conn.ReadLine()
		if err != nil {
			log.Println("Failed reading line", err)
			break
		}
		
		err = conn.PrintfLine(cmd)
		if err != nil {
			log.Println("Failed writing line", err)
			break
		}
	}
}

func listenTCPClients() {
	
	ipAddr, err := net.ResolveIPAddr("tcp4", "127.0.0.1")
	if err != nil {
		log.Fatal("Failed to resolve", err)
	}

	addr := net.TCPAddr{ipAddr.IP, 3535}

	listener, err := net.ListenTCP("tcp4", &addr)
	if err != nil {
		log.Fatal("Failed to listen", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept", err)
		}

		protoConn := textproto.NewConn(conn)
		go serveTCP(protoConn)
	}
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

	go listenTCPClients()

	http.Handle("/", http.HandlerFunc(ServePage))
	http.Handle("/ws", websocket.Handler(ServeWS))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}
