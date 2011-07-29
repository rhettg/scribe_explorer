package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"http"
	"websocket"
	"log"
	"net"
	"flag"
)

type LineReader interface {
	ReadLine() (line []byte, isPrefix bool, err os.Error)
}

var (
	rangerStream *DataStream
)

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

	defer func() {
        if err := recover(); err != nil {
            log.Println("ServeWS failed:", err)
        }
    }()

	ServeStream(jsonStream)
}

func serveTCP(conn net.Conn) {
	defer conn.Close()
	
	defer func() {
        if err := recover(); err != nil {
            log.Println("serveTCP failed:", err)
        }
    }()

	jsonStream := NewJSONConn(conn)

	ServeStream(jsonStream)
}

func ServeStream(stream *JSONConn) {
	// Create a new channel to receive data on
	dataChan := make(chan JSONData, 16)
	request := new(SubscribeRequest)
	request.dataChan = dataChan
	rangerStream.subscribeChan <- request
	
	defer func() {rangerStream.unsubscribeChan <- request}()

	// Get our query from the client
	query, err := stream.ReadJSON()
	if err != nil {
		log.Printf("Failed to read from client", err)
		return
	}

	displayFields := []string{}
	aggregators := []Aggregator{}

	for _, fieldValue := range query.(map[string] interface{})["fields"].([]interface{}) {
		aggregator := ParseAggregatorStatement(fieldValue.(string))
		if aggregator != nil {
			aggregators = append(aggregators, aggregator)
			log.Printf("Parsed to aggregator: ", aggregator.String())
		}else {
			displayFields = append(displayFields, fieldValue.(string))
			log.Printf("Field: ", fieldValue)
		}
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

		for _, aggregator := range aggregators {
			outputMap[aggregator.String()] = aggregator.Push(data)
		}

		err := stream.WriteJSON(outputMap)
		if err != nil {
			log.Printf("Failed to write", err)
			return
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

		//protoConn := textproto.NewConn(conn)
		go serveTCP(conn)
	}
}

var aggregator = flag.String("e", "dev", "One of {dev, stagea, stagex, prod}")

func main() {
	log.Println("Starting up")

	flag.Parse()
	streamHost := fmt.Sprintf("scribe-%s.local.yelpcorp.com:3535", *aggregator)
	log.Println("Connecting to ", streamHost)

	rangerStream = NewDataStream("ranger", streamHost)

	go listenTCPClients()

	http.Handle("/", http.HandlerFunc(ServePage))
	http.Handle("/ws", websocket.Handler(ServeWS))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}
