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
	streamHost string
	scribeStreams map[string] *DataStream

)

func init() {
	scribeStreams = make(map[string] *DataStream)
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

// type ScribeQuery struct {
// 	fields []string
// 	filters []string
//  logName string
// }

func ServeStream(stream *JSONConn) {
	// Get our query from the client
	query, err := stream.ReadJSON()
	if err != nil {
		log.Printf("Failed to read from client", err)
		return
	}

	// Find the stream
	logName := query.(map[string] interface{})["logName"].(string)
	log.Printf("Subscribing to log", logName)
	
	scribeStream := StreamByName(logName)
	
	// Create a new channel to receive data on
	dataChan := make(chan JSONData, 16)
	request := new(SubscribeRequest)
	request.dataChan = dataChan
	scribeStream.subscribeChan <- request
	
	defer func() {scribeStream.unsubscribeChan <- request}()


	displayFields := []Expression{}
	for _, fieldValue := range query.(map[string] interface{})["fields"].([]interface{}) {
		aggregator, err := Parse(fieldValue.(string))
		if err != nil {
			log.Printf("Couldn't parse expression %v: %v", fieldValue, err)
		}else {
			displayFields = append(displayFields, aggregator)
			log.Printf("Parsed to aggregator: %v", aggregator.String())
		}
	}

	filterPredicates := []Expression{}
	for _, statement := range query.(map[string] interface{})["filters"].([]interface{}) {
		log.Printf("Statement: ", statement)
		expr, err := Parse(statement.(string))
		if err != nil {
			log.Printf("Couldn't parse statement \"%s\": %v", statement, err)	
		}else {
			filterPredicates = append(filterPredicates, expr)
		}
	}


	for {
		data := <-dataChan
		
		if passes, err := PassesAllFilters(data, filterPredicates); !passes {
			if err != nil {
				log.Printf("Got error evaluating predicates: %v", err)
			}
			continue
		}
		outputMap := map[string] interface{}{}
		
		for _, fieldValue := range displayFields {
			result, err := fieldValue.Evaluate(data)
			if err == nil {
				outputMap[fieldValue.String()] = result
			} else {
				outputMap[fieldValue.String()] = nil	
				log.Printf("Got error '%v' evaluating field '%v'", err, fieldValue)
			}
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

func StreamByName(name string) (stream *DataStream) {
	if stream, ok := scribeStreams[name]; ok {
		return stream
	}
	
	scribeStreams[name] = NewDataStream(name, streamHost)
	return scribeStreams[name]
}

var aggregator = flag.String("e", "dev", "One of {dev, stagea, stagex, prod}")

func main() {
	log.Println("Starting up")

	flag.Parse()
	streamHost = fmt.Sprintf("scribe-%s.local.yelpcorp.com:3535", *aggregator)
	log.Println("Connecting to ", streamHost)

	go listenTCPClients()

	http.Handle("/", http.HandlerFunc(ServePage))
	http.Handle("/ws", websocket.Handler(ServeWS))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}
