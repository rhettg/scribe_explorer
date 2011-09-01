package main
import (
	"net"
	"bufio"
	"io"
	"json"
	"log"
	"os"
)

type SubscribeRequest struct {
	dataChan chan JSONData
	id int
}

type DataStream struct {
	name string
	connectString string
	
	rawStream io.ReadWriteCloser  // Raw io stream of data
	ioStream *bufio.Reader  // Our buffered view of our data stream

	allChannels [](chan JSONData)
	nextChanNdx int
	
	SubscribeChan chan *SubscribeRequest
	UnsubscribeChan chan *SubscribeRequest
	ReceiveNextChan chan (chan JSONData)  // Channel to pull channels out of for iterating through our allChannel set
	
}

func NewDataStream(name string, connectString string) (stream *DataStream) {
	stream = new(DataStream)
	stream.name = name
	stream.connectString = connectString

	stream.allChannels = make([](chan JSONData), 0, 64)

	stream.SubscribeChan = make(chan *SubscribeRequest)
	stream.UnsubscribeChan = make(chan *SubscribeRequest)
	stream.ReceiveNextChan = make(chan (chan JSONData))
	
	go stream.acceptChannels()
	return
}

func (stream *DataStream) acceptChannels() {
	
	nextChan := stream.findNextChan()

	for {
		select {
			case channelRequest := <-stream.SubscribeChan:
				stream.subscribe(channelRequest)
			case channelRequest := <-stream.UnsubscribeChan:
				stream.unsubscribe(channelRequest)
			case stream.ReceiveNextChan<- nextChan:
				nextChan = stream.findNextChan()
		}
	}
}

func (stream *DataStream) subscribe(request *SubscribeRequest) {
	request.id = -1
	for ndx, value := range(stream.allChannels) {
		if value == nil {
			stream.allChannels[ndx] = request.dataChan
			request.id = ndx
			break
		}
	}
	if request.id < 0 {
		stream.allChannels = append(stream.allChannels, request.dataChan)
		request.id = (len(stream.allChannels) - 1)
	}
	log.Printf("Adding new channel %d to data stream", request.id, stream.name)

	// If we are not yet streaming data, we should be
	if stream.ioStream == nil {
		stream.createIOStream()
		go stream.streamData()
	}
}

func (stream *DataStream) unsubscribe(request *SubscribeRequest) {
	log.Println("Dropping channel", request.id)
	stream.allChannels[request.id] = nil
}

/*
  Find next channel for our stream processors

  Updates the stream.nextChanNdx attribute, which is an index into allChannels
  */
func (stream *DataStream) findNextChan() chan JSONData {
	for stream.nextChanNdx++; stream.nextChanNdx < len(stream.allChannels); stream.nextChanNdx++ {
		if stream.allChannels[stream.nextChanNdx] != nil {
			return stream.allChannels[stream.nextChanNdx]
		}
	}

	stream.nextChanNdx = -1
	return nil
}

func (stream *DataStream) streamData() {
	// The first receive channel is always nil
	dataChannel := <-stream.ReceiveNextChan
	if dataChannel != nil {
		log.Fatal("whaaaa?")
	}
	
	for {
		line, isPrefix, err := stream.ioStream.ReadLine()
		if err != nil {
			if err == os.EOF {
				break
			}
			log.Printf("Failed on line stream", err)
			break
		}
		if isPrefix {
			log.Printf("PREFIX!! Skipping line.")
			continue
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
		sent := false
		for {
			dataChannel = <-stream.ReceiveNextChan	
			if dataChannel == nil {
				break
			}

			select {
				case dataChannel <- data:
					sent = true
				default:
					log.Println("Dropping data to channel")
			}
		}		

		/* There are no dataChannel's left open, we can close the stream 
		   TODO: I'm suspicious of race conditions here. Closing the stream should probably be handled by the 
		          acceptChannel goroutine.
		*/
		if !sent {
			log.Printf("Closing data stream for %s", stream.name)
			stream.rawStream.Close()
			stream.rawStream = nil;
			stream.ioStream = nil;
			break
		}
	}
	log.Printf("All done with data stream %s", stream.name)
}


func (stream *DataStream) createIOStream() {
	conn, err := net.Dial("tcp4", stream.connectString)
	if err != nil {
		log.Fatal("Failed to open", err)
	}

	_, err = conn.Write([]uint8(stream.name + "\n"))
	if err != nil {
		log.Fatal("Failed to send cmd", err)
	}

	stream.rawStream = conn
	stream.ioStream, err = bufio.NewReaderSize(conn, 1024*32)
	if err != nil {
		log.Fatal("Failed to create reader", err)
	}
}
